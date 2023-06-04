// -- sideline-header-v0.1 --
// Copyright (c) Sideline Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package sideline

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gosimple/slug"
	sidelinepb "github.com/rocky-linux/sideline/pb"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/ulikunitz/xz"
	"golang.org/x/sync/errgroup"
)

var (
	ErrCouldNotDetectSource0   = errors.New("could not detect source0")
	ErrCouldNotFindSource0File = errors.New("could not find source0 file")
)

var (
	indexRegex  = regexp.MustCompile(`(?m)<a href="(.+\.rpm)">.+\.rpm</a>\s+(.+\d)\s+(\d+)`)
	nvrRegex    = regexp.MustCompile("^(\\S+)-([\\w~.]+)-(\\w+(?:\\.[\\w+]+)+?)(?:\\.(\\w+))?(?:\\.rpm)?$")
	gitLogRegex = regexp.MustCompile(`(\w{40}) (.*) (\d{4}-\d{2}-\d{2}) (.*)`)
)

type Response struct {
	Srpmproc     *srpmprocpb.Cfg
	PatchName    string
	PatchContent string
}

type Commit struct {
	Hash    string
	Message string
	Author  string
	Date    string
}

type BackportResponse struct {
	AffectingCommits []*Commit
}

type UpstreamRepo struct {
	Fetch   *FetchRepo
	Repo    *git.Repository
	Remote  *git.Remote
	TempDir *string
}

type UpstreamRepos struct {
	Target      *UpstreamRepo
	CompareWith *UpstreamRepo
}

type FetchRepo struct {
	URL     string
	Tag     *string
	Branch  *string
	Depth   int32
	UseOsfs bool
	Mode    sidelinepb.Upstream_CompareMode
}

// fetchUnpackDistributionSource clones the distribution source from git
// in-memory and does unpacking according to the distribution.
// Currently source RPMs are supported. Instead of cloning the SCM source,
// we're going the way of source RPMs because of patches that has to be applied.
// If we do an SCM clone and fetch the tarballs, we would also need to do
// rpmbuild to take care of the patches.
// We may add support for SCM clones as well as other distributions in the
// future, but let's keep it simple for now.
func fetchUnpackDistributionSource(srpmUrl string) (*git.Repository, *rpm.Package, error) {
	var body io.ReadCloser

	// Fetch the SRPM file
	if strings.HasPrefix(srpmUrl, "file://") {
		// Allow local filesystem pull too
		// In RESF infra we may rely on the internal NFS shares
		// to reach the necessary SRPMs
		log.Printf("Fetching %s from local filesystem", srpmUrl)
		f, err := os.Open(strings.TrimPrefix(srpmUrl, "file://"))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open %s: %v", srpmUrl, err)
		}
		body = f
	} else {
		log.Printf("Fetching %s", srpmUrl)
		resp, err := http.Get(srpmUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch source RPM: %v", err)
		}
		body = resp.Body
	}
	defer func() {
		err := body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}()

	// Read the SRPM file to find Source0
	// We're going to unpack Source0, lay upstream changes on top
	// and create patches that srpmproc can use
	rpmFile, err := rpm.Read(body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read SRPM: %v", err)
	}

	// Assume Source0 is the first file in the RPM
	// Because of FIFO, the first file should be last element in the list
	sources := rpmFile.Source()
	source0 := sources[len(sources)-1]

	// Let's have some failsafe around this practice
	// We're going to assume that the source0 is a tarball
	// If not a tarball, we need to look for a tarball
	if !strings.Contains(source0, ".tar") {
		for _, source := range sources {
			if strings.Contains(source, ".tar") {
				source0 = source
				break
			}
		}
	}
	if !strings.Contains(source0, ".tar") {
		return nil, nil, ErrCouldNotDetectSource0
	}
	log.Printf("Found source0: %s", source0)

	var payloadReader io.Reader

	compression := rpmFile.PayloadCompression()
	switch compression {
	case "xz":
		payloadReader, err = xz.NewReader(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create xz reader: %v", err)
		}
	case "gz",
		"gzip":
		payloadReader, err = gzip.NewReader(body)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
	}

	format := rpmFile.PayloadFormat()
	// make sure it's cpio
	if format != "cpio" {
		return nil, nil, fmt.Errorf("unsupported payload format: %s", format)
	}

	var source0Reader io.Reader

	cpioReader := cpio.NewReader(payloadReader)
	for {
		hdr, err := cpioReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read cpio header: %v", err)
		}

		if hdr.Name == source0 {
			source0Reader = cpioReader
			break
		}
	}

	if source0Reader == nil {
		return nil, nil, ErrCouldNotFindSource0File
	}

	// Now let's unpack the source0 in memory
	// and create a git repository by its contents.
	// This way when the changes from upstream are laid on top,
	// we can create patches that can be applied to the source0
	s := memory.NewStorage()
	fs := memfs.New()
	g, err := git.Init(s, fs)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init git repository: %v", err)
	}

	w, err := g.Worktree()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get worktree: %v", err)
	}

	// Let's read the tar content and "unpack"
	// First let's uncompress if any
	switch c := filepath.Ext(source0); c {
	case ".gz":
		gzReader, err := gzip.NewReader(source0Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		source0Reader = gzReader
	case ".xz":
		xzReader, err := xz.NewReader(source0Reader)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create xz reader: %v", err)
		}
		source0Reader = xzReader
	default:
		return nil, nil, fmt.Errorf("unsupported source0 compression: %s", c)
	}
	tarReader := tar.NewReader(source0Reader)

	// Now let's unpack the tar content into the git repository
	// Source0's usually have a single directory that contains the source
	for {
		hdr, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read tar header: %v", err)
		}

		nameSplit := strings.SplitN(hdr.Name, "/", 2)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if len(nameSplit) <= 1 {
				continue
			}
			err = w.Filesystem.MkdirAll(nameSplit[1], 0755)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			f, err := w.Filesystem.Create(nameSplit[1])
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create file: %v", err)
			}
			_, err = io.Copy(f, tarReader)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to copy file: %v", err)
			}
			err = f.Close()
			if err != nil {
				return nil, nil, fmt.Errorf("failed to close file: %v", err)
			}
		}
	}

	// Create an initial commit to use as the base for the patches
	_, err = w.Add(".")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add files: %v", err)
	}
	h, err := w.Commit("Initial commit", &git.CommitOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to commit: %v", err)
	}
	log.Printf("Created initial commit %s", h.String())

	// Return the repo back
	return g, rpmFile, nil
}

func fetchUpstream(upstream *FetchRepo, compareWith *FetchRepo) (*UpstreamRepos, error) {
	var eg errgroup.Group

	var g *git.Repository
	eg.Go(func() error {
		if upstream.Tag != nil {
			log.Printf("Using upstream repo: %s (tag %s)", upstream.URL, *upstream.Tag)
		} else {
			log.Printf("Using upstream repo: %s (branch %s)", upstream.URL, *upstream.Branch)
		}
		s := memory.NewStorage()
		fs := memfs.New()

		var refName plumbing.ReferenceName
		if upstream.Tag != nil {
			refName = plumbing.NewTagReferenceName(*upstream.Tag)
		} else {
			refName = plumbing.NewBranchReferenceName(*upstream.Branch)
		}
		var err error
		g, err = git.Clone(s, fs, &git.CloneOptions{
			URL:           upstream.URL,
			ReferenceName: refName,
			SingleBranch:  true,
			Depth:         int(upstream.Depth),
			Progress:      os.Stdout,
			Tags:          git.NoTags,
		})
		if err != nil {
			return fmt.Errorf("failed to clone upstream: %v", err)
		}

		return nil
	})

	var compareWithRepo *git.Repository
	var tempDir *string
	eg.Go(func() error {
		if compareWith != nil {
			log.Printf("Using compare repo: %s (tag %s)", compareWith.URL, *compareWith.Tag)

			var s2 storage.Storer
			var fs2 billy.Filesystem
			if compareWith.UseOsfs {
				tmpDir, err := ioutil.TempDir("", "sideline-")
				if err != nil {
					return fmt.Errorf("failed to create temp dir: %v", err)
				}
				defer func() {
					if compareWithRepo == nil {
						_ = os.RemoveAll(tmpDir)
					}
				}()

				log.Printf("Using temp dir %s for compare storage", tmpDir)

				fs2 = osfs.New(tmpDir)

				dot, _ := fs2.Chroot(git.GitDirName)
				s2 = filesystem.NewStorage(dot, cache.NewObjectLRUDefault())
				tempDir = &tmpDir
			} else {
				fs2 = memfs.New()
				s2 = memory.NewStorage()
			}

			g2, err := git.Init(s2, fs2)
			if err != nil {
				return fmt.Errorf("failed to init git repository: %v", err)
			}
			_, err = g2.CreateRemote(&config.RemoteConfig{
				Name: "origin",
				URLs: []string{compareWith.URL},
			})
			if err != nil {
				return fmt.Errorf("failed to create remote: %v", err)
			}

			refspecs := []config.RefSpec{"+refs/heads/*:refs/remotes/origin/*"}
			if compareWith.Mode != sidelinepb.Upstream_Unknown {
				switch compareWith.Mode {
				case sidelinepb.Upstream_KernelTag:
					targetV := (*upstream.Tag)[0:2]
					refspecTarget := config.RefSpec(fmt.Sprintf("+refs/tags/%s.*:refs/remotes/origin/tags/%[1]s.*", targetV))
					refspecCompare := config.RefSpec(fmt.Sprintf("+refs/tags/%s:refs/remotes/origin/tags/%[1]s", *compareWith.Tag))
					refspecs = []config.RefSpec{refspecTarget, refspecCompare}
				}
			}
			log.Printf("Using refspecs for compare repo: %v", refspecs)
			err = g2.Fetch(&git.FetchOptions{
				RemoteName: "origin",
				RefSpecs:   refspecs,
				Progress:   os.Stdout,
				Tags:       git.AllTags,
				Force:      true,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch upstream: %v", err)
			}

			g2w, err := g2.Worktree()
			if err != nil {
				return fmt.Errorf("failed to get worktree: %v", err)
			}

			err = g2w.Checkout(&git.CheckoutOptions{
				Branch: plumbing.NewTagReferenceName(*compareWith.Tag),
				Force:  true,
			})
			if err != nil {
				return fmt.Errorf("failed to checkout in compare: %v", err)
			}

			compareWithRepo = g2
		}

		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	ret := &UpstreamRepos{
		Target: &UpstreamRepo{
			Fetch: upstream,
			Repo:  g,
		},
		CompareWith: nil,
	}
	if compareWithRepo != nil {
		ret.CompareWith = &UpstreamRepo{
			Fetch:   compareWith,
			Repo:    compareWithRepo,
			TempDir: tempDir,
		}
	}
	return ret, nil
}

func recursivelyDeletePath(to billy.Filesystem, path string) error {
	read, err := to.ReadDir(path)
	if err != nil {
		return fmt.Errorf("could not read dir: %v", err)
	}

	for _, fi := range read {
		fullPath := filepath.Join(path, fi.Name())

		if fi.IsDir() {
			err := recursivelyDeletePath(to, fullPath)
			if err != nil {
				return err
			}
			err = to.Remove(fullPath)
			if err != nil {
				return fmt.Errorf("could not remove dir: %v", err)
			}
		} else {
			err := to.Remove(fullPath)
			if err != nil {
				return fmt.Errorf("could not remove file: %v", err)
			}
		}
	}

	return nil
}

func recursivelyCopyPath(from billy.Filesystem, to *git.Worktree, path string) error {
	read, err := from.ReadDir(path)
	if err != nil {
		return fmt.Errorf("could not read dir: %v", err)
	}

	for _, fi := range read {
		fullPath := filepath.Join(path, fi.Name())

		if fi.IsDir() {
			_ = to.Filesystem.MkdirAll(fullPath, 0755)
			err := recursivelyCopyPath(from, to, fullPath)
			if err != nil {
				return err
			}
		} else {
			_ = to.Filesystem.Remove(fullPath)

			f, err := to.Filesystem.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fi.Mode())
			if err != nil {
				return fmt.Errorf("could not open file: %v", err)
			}

			oldFile, err := from.Open(fullPath)
			if err != nil {
				return fmt.Errorf("could not open from file: %v", err)
			}

			_, err = io.Copy(f, oldFile)
			if err != nil {
				return fmt.Errorf("could not copy from oldFile to new: %v", err)
			}

			_, err = to.Add(fullPath)
			if err != nil {
				return fmt.Errorf("could not add file: %v", err)
			}
		}
	}

	return nil
}

func readFileBillyFs(fs billy.Filesystem, path string) ([]byte, error) {
	f, err := fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %v", err)
	}
	defer func(f billy.File) {
		_ = f.Close()
	}(f)

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	return b, nil
}

func writeFileBillyFs(fs billy.Filesystem, path string, data []byte) error {
	f, err := fs.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("could not open file: %v", err)
	}
	defer func(f billy.File) {
		_ = f.Close()
	}(f)

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}

	return nil
}

func backportUpstreamChanges(distro *git.Repository, repos *UpstreamRepos, changes []*sidelinepb.Changes) (*BackportResponse, error) {
	log.Printf("Backporting upstream changes")
	res := &BackportResponse{
		AffectingCommits: []*Commit{},
	}

	distroW, err := distro.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get distro worktree: %v", err)
	}

	upstreamW, err := repos.Target.Repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get upstream worktree: %v", err)
	}

	// First do recursive overrides
	for i, change := range changes {
		log.Printf("==> Executing change %d", i)
		logNested := log.New(os.Stderr, "\t==> ", log.LstdFlags|log.Lmsgprefix)
		// Always first upstream to distro operations.
		// Currently, only "recursive_path"
		for _, recursivePath := range change.RecursivePath {
			logNested.Printf("Recursively copying %s", recursivePath)
			logNested = log.New(os.Stderr, "\t\t==> ", log.LstdFlags|log.Lmsgprefix)

			// Check that the recursive path exists in the upstream and is a directory
			fi, err := upstreamW.Filesystem.Stat(recursivePath)
			if err != nil {
				return nil, fmt.Errorf("failed to read directory: %v", err)
			}
			if !fi.IsDir() {
				return nil, fmt.Errorf("recursive path %s is not a directory in upstream", recursivePath)
			}

			// For a recursive copy, we can just check commits that touch the path as a whole
			// This is if there is a compare tag set, and we want the affecting commits
			if repos.CompareWith != nil {
				logNested.Printf("Checking commits that touch %s", recursivePath)
				compareHead, err := repos.CompareWith.Repo.Head()
				if err != nil {
					return nil, fmt.Errorf("failed to get compare head: %v", err)
				}
				compareCommit, err := repos.CompareWith.Repo.CommitObject(compareHead.Hash())
				if err != nil {
					return nil, fmt.Errorf("failed to get compare commit: %v", err)
				}
				logNested.Printf("Using compare commit %s", compareCommit.Hash)

				targetHead, err := repos.Target.Repo.Head()
				if err != nil {
					return nil, fmt.Errorf("failed to get target head: %v", err)
				}
				targetCommit, err := repos.Target.Repo.CommitObject(targetHead.Hash())
				if err != nil {
					return nil, fmt.Errorf("failed to get target commit: %v", err)
				}
				logNested.Printf("Using target commit %s", targetCommit.Hash)

				// For now, let's shell out to git for the log operation
				// go-git unfortunately doesn't support the exact use case we're after
				// Instead of making it more complicated, we'll just shell out to git
				// git log 94710cac0ef4ee177a63b5227664b38c95bbf703...f443e374ae131c168a065ea1748feac6b2e76613 --pretty=format:"%H %an %ad %s" --date=short -- drivers/net/ethernet/google
				var output bytes.Buffer
				cmd := exec.Command("git", "log", fmt.Sprintf("%s...%s", compareCommit.Hash, targetCommit.Hash), "--pretty=format:%H %an %ad %s", "--date=short", "--", recursivePath)
				cmd.Dir = *repos.CompareWith.TempDir
				cmd.Stdout = &output
				err = cmd.Run()
				if err != nil {
					return nil, fmt.Errorf("failed to run git log: %v", err)
				}

				strOutput := output.String()
				strLines := strings.Split(strOutput, "\n")
				for _, strLine := range strLines {
					match := gitLogRegex.FindStringSubmatch(strLine)
					hash := match[1]
					author := match[2]
					date := match[3]
					message := match[4]

					res.AffectingCommits = append(res.AffectingCommits, &Commit{
						Hash:    hash,
						Message: message,
						Author:  author,
						Date:    date,
					})
				}

				logNested.Printf("Added changelog for %d commits", len(res.AffectingCommits))
			}

			// First delete the path from distro worktree
			err = recursivelyDeletePath(distroW.Filesystem, recursivePath)
			if err != nil {
				log.Printf("failed to remove path: %v, but continuing", err)
			}

			// Then re-create the directory
			err = distroW.Filesystem.MkdirAll(recursivePath, 0755)
			if err != nil {
				return nil, fmt.Errorf("failed to create directory: %v", err)
			}
			_, err = distroW.Add(recursivePath)
			if err != nil {
				return nil, fmt.Errorf("failed to add path: %v", err)
			}

			err = recursivelyCopyPath(upstreamW.Filesystem, distroW, recursivePath)
			if err != nil {
				return nil, fmt.Errorf("failed to copy path %s: %v", recursivePath, err)
			}
		}

		// Now run file change operations
		for _, fileChange := range change.FileChange {
			log.Printf("\t==> Making changes to %s", fileChange.Path)

			// Search and replace first (and only right now)
			for _, searchReplace := range fileChange.SearchReplace {
				fileContents, err := readFileBillyFs(distroW.Filesystem, fileChange.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to read file: %v", err)
				}

				if !strings.HasPrefix(searchReplace.Find, "(!regex)") {
					fileContents = []byte(strings.Replace(string(fileContents), searchReplace.Find, searchReplace.Replace, int(searchReplace.N)))
				} else {
					regx, err := regexp.Compile(strings.TrimPrefix(searchReplace.Find, "(!regex)"))
					if err != nil {
						return nil, fmt.Errorf("failed to compile regex: %v", err)
					}
					fileContents = []byte(regx.ReplaceAllString(string(fileContents), searchReplace.Replace))
				}
				err = writeFileBillyFs(distroW.Filesystem, fileChange.Path, fileContents)
				if err != nil {
					return nil, fmt.Errorf("failed to write file: %v", err)
				}

				_, err = distroW.Add(fileChange.Path)
				if err != nil {
					return nil, fmt.Errorf("failed to add file: %v", err)
				}
			}
		}
	}

	// show status
	status, err := distroW.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %v", err)
	}
	log.Printf("Successfully processed:\n%s", status)

	statusLines := strings.Split(status.String(), "\n")
	for _, line := range statusLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "D") {
			path := strings.TrimPrefix(trimmed, "D ")
			_, err := distroW.Remove(path)
			if err != nil {
				return nil, fmt.Errorf("could not delete extra file %s: %v", path, err)
			}
		}
	}

	return res, nil
}

func validateCfg(cfg *sidelinepb.Configuration) error {
	switch cfg.Preset {
	case "rocky8":
		cfg.PresetOverride = &sidelinepb.PresetOverride{
			Override: &sidelinepb.PresetOverride_BaseUrl{
				BaseUrl: "https://dl.rockylinux.org/pub/rocky/8",
			},
		}
	case "rocky9":
		cfg.PresetOverride = &sidelinepb.PresetOverride{
			Override: &sidelinepb.PresetOverride_BaseUrl{
				BaseUrl: "https://dl.rockylinux.org/pub/rocky/9",
			},
		}
	case "manual":
		if cfg.PresetOverride == nil || cfg.PresetOverride.Override == nil {
			return fmt.Errorf("preset override is required for manual")
		}

		switch c := cfg.PresetOverride.Override.(type) {
		case *sidelinepb.PresetOverride_BaseUrl:
			if c.BaseUrl == "" {
				return fmt.Errorf("base url cannot be empty")
			}
		case *sidelinepb.PresetOverride_FullUrl:
			if c.FullUrl == "" {
				return fmt.Errorf("full url cannot be empty")
			}
		}
	default:
		return fmt.Errorf("unknown preset: %s", cfg.Preset)
	}

	if len(strings.SplitN(cfg.Package, "//", 2)) != 2 {
		return fmt.Errorf("package must be in the form of 'repo//nvra'")
	}

	return nil
}

func tryUrl(attemptUrl string) (bool, error) {
	resp, err := http.Get(attemptUrl)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		return false, nil
	}

	return true, nil
}

func getHttpBody(urlToGet string) (string, error) {
	resp, err := http.Get(urlToGet)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get %s: %s", urlToGet, resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func constructSrpmUrl(cfg *sidelinepb.Configuration) (string, error) {
	var baseUrl string

	switch c := cfg.PresetOverride.Override.(type) {
	case *sidelinepb.PresetOverride_FullUrl:
		return c.FullUrl, nil
	case *sidelinepb.PresetOverride_BaseUrl:
		baseUrl = c.BaseUrl
	}

	// Try to fetch the latest SRPM
	repoPackage := strings.SplitN(cfg.Package, "//", 2)
	repo := repoPackage[0]
	pkg := repoPackage[1]

	baseUrl = fmt.Sprintf("%s/%s/source/tree/Packages", baseUrl, repo)

	// Handle filesystem based repos here
	if strings.HasPrefix(baseUrl, "file://") {
		return "", fmt.Errorf("file:// repos are unimplemented")
	}

	// Some repos use categorization based on the first character in the file name (Rocky does)
	initialUrl := fmt.Sprintf("%s/%s", baseUrl, strings.ToLower(string(pkg[0])))
	tryInitial, err := tryUrl(initialUrl)
	if err != nil {
		return "", fmt.Errorf("failed to try initial url: %v", err)
	}
	if tryInitial {
		baseUrl = initialUrl
	}

	// Get the directory listing
	directoryListingHtml, err := getHttpBody(baseUrl)
	if err != nil {
		return "", fmt.Errorf("failed to get directory listing: %v", err)
	}

	rpmEntries := indexRegex.FindAllStringSubmatch(directoryListingHtml, -1)
	if len(rpmEntries) == 0 {
		return "", fmt.Errorf("failed to find any RPMs in %s", baseUrl)
	}

	// Find the latest SRPM
	var latestSrpm string
	var latestDate *time.Time
	for _, entry := range rpmEntries {
		if strings.HasSuffix(entry[1], ".src.rpm") {
			// Parse RPM information
			nvr := nvrRegex.FindStringSubmatch(entry[1])
			if len(nvr) == 0 {
				log.Printf("failed to parse nvr from %s", entry[1])
				continue
			}
			if nvr[1] != pkg {
				continue
			}

			date, err := time.Parse("02-Jan-2006 15:04", entry[2])
			if err != nil {
				return "", fmt.Errorf("failed to parse date: %v", err)
			}

			if latestDate == nil {
				latestDate = &date
				latestSrpm = entry[1]
				continue
			}

			if date.After(*latestDate) {
				latestSrpm = entry[1]
				latestDate = &date
			}
		}
	}

	return fmt.Sprintf("%s/%s", baseUrl, latestSrpm), nil
}

func Run(cfg *sidelinepb.Configuration) (*Response, error) {
	if err := validateCfg(cfg); err != nil {
		return nil, err
	}

	srpmUrl, err := constructSrpmUrl(cfg)
	if err != nil {
		return nil, err
	}

	distro, srpm, err := fetchUnpackDistributionSource(srpmUrl)
	if err != nil {
		return nil, err
	}

	upstreamFetchRepo := &FetchRepo{
		URL:   cfg.Upstream.GetGit(),
		Depth: cfg.Upstream.Depth,
	}
	if x := cfg.Upstream.GetTag(); x != "" {
		upstreamFetchRepo.Tag = &x
	} else if x := cfg.Upstream.GetBranch(); x != "" {
		upstreamFetchRepo.Branch = &x
	}
	var compareWithFetchRepo *FetchRepo
	if cfg.Upstream.GetCompareWithTag() != nil {
		compareWithTag := cfg.Upstream.GetCompareWithTag().GetValue()
		compareWithFetchRepo = &FetchRepo{
			URL:   cfg.Upstream.GetGit(),
			Tag:   &compareWithTag,
			Depth: cfg.Upstream.Depth,
			// todo(mustafa): support in memory `git log`
			UseOsfs: true,
			Mode:    cfg.Upstream.CompareMode,
		}
	}
	upstream, err := fetchUpstream(upstreamFetchRepo, compareWithFetchRepo)
	if err != nil {
		return nil, err
	}

	// Delete the temp dir after returning
	if upstream.CompareWith != nil && upstream.CompareWith.TempDir != nil {
		defer func() {
			_ = os.RemoveAll(*upstream.CompareWith.TempDir)
		}()
	}

	backportRes, err := backportUpstreamChanges(distro, upstream, cfg.Changes)
	if err != nil {
		return nil, err
	}

	distroW, err := distro.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get distro worktree: %v", err)
	}

	objects, err := distro.CommitObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit objects: %v", err)
	}
	initialObject, _ := objects.Next()

	commitMsg := fmt.Sprintf("Backport changes from %s (tag: %s)", cfg.Upstream.GetGit(), cfg.Upstream.GetTag())
	commitHash, err := distroW.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  cfg.ContactName,
			Email: cfg.ContactEmail,
			When:  time.Now(),
		},
	})
	commit, err := distro.CommitObject(commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %v", err)
	}
	genPatch, err := initialObject.Patch(commit)
	if err != nil {
		return nil, fmt.Errorf("failed to get patch: %v", err)
	}

	commitMsgs := []string{commitMsg}
	// If we have detected the affecting commits, add to commit messages
	if len(backportRes.AffectingCommits) > 0 {
		var newAffectingCommits []string
		for _, affectingCommit := range backportRes.AffectingCommits {
			newAffectingCommits = append(newAffectingCommits, fmt.Sprintf("%s - %s - %s - %s", affectingCommit.Hash, affectingCommit.Author, affectingCommit.Date, affectingCommit.Message))
		}
		commitMsgs = append(commitMsgs, newAffectingCommits...)
	}

	patchName := fmt.Sprintf("%d-%s.patch", len(srpm.Patch())+1, slug.Make(commitMsg))
	response := &Response{
		Srpmproc: &srpmprocpb.Cfg{
			Add: []*srpmprocpb.Add{
				{
					Source: &srpmprocpb.Add_File{
						File: fmt.Sprintf("SIDELINE/_supporting/%s", patchName),
					},
				},
			},
			SpecChange: &srpmprocpb.SpecChange{
				File: []*srpmprocpb.SpecChange_FileOperation{
					{
						Name: patchName,
						Type: srpmprocpb.SpecChange_FileOperation_Patch,
						Mode: &srpmprocpb.SpecChange_FileOperation_Add{
							Add: true,
						},
					},
				},
				Changelog: []*srpmprocpb.SpecChange_ChangelogOperation{
					{
						AuthorName:  cfg.ContactName,
						AuthorEmail: cfg.ContactEmail,
						Message:     commitMsgs,
					},
				},
			},
		},
		PatchName:    patchName,
		PatchContent: genPatch.String(),
	}

	if cfg.ApplyPatch != nil {
		switch x := cfg.ApplyPatch.Patch.(type) {
		case *sidelinepb.ApplyPatch_AutoPatch:
			response.Srpmproc.SpecChange.File[0].AddToPrep = x.AutoPatch.AddToPrep
			response.Srpmproc.SpecChange.File[0].NPath = x.AutoPatch.NPath
		case *sidelinepb.ApplyPatch_Custom:
			response.Srpmproc.SpecChange.File = append(response.Srpmproc.SpecChange.File, x.Custom.File...)
			response.Srpmproc.SpecChange.Changelog = append(response.Srpmproc.SpecChange.Changelog, x.Custom.Changelog...)
			response.Srpmproc.SpecChange.Append = append(response.Srpmproc.SpecChange.Append, x.Custom.Append...)
			response.Srpmproc.SpecChange.NewField = append(response.Srpmproc.SpecChange.NewField, x.Custom.NewField...)
			response.Srpmproc.SpecChange.SearchAndReplace = append(response.Srpmproc.SpecChange.SearchAndReplace, x.Custom.SearchAndReplace...)
			response.Srpmproc.SpecChange.DisableAutoAlign = x.Custom.DisableAutoAlign

			for _, sar := range x.Custom.SearchAndReplace {
				sar.Replace = strings.Replace(sar.Replace, "%%patch_name%%", patchName, -1)
			}
		}
	}

	return response, nil
}
