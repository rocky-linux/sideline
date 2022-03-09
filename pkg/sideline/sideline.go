package sideline

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/cavaliergopher/cpio"
	"github.com/cavaliergopher/rpm"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gosimple/slug"
	sidelinepb "github.com/rocky-linux/sideline/pb"
	srpmprocpb "github.com/rocky-linux/srpmproc/pb"
	"github.com/ulikunitz/xz"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var (
	ErrCouldNotDetectSource0   = errors.New("could not detect source0")
	ErrCouldNotFindSource0File = errors.New("could not find source0 file")
)

var (
	indexRegex = regexp.MustCompile(`(?m)<a href="(.+\.rpm)">.+\.rpm</a>\s+(.+\d)\s+(\d+)`)
	nvrRegex   = regexp.MustCompile("^(\\S+)-([\\w~.]+)-(\\w+(?:\\.[\\w+]+)+?)(?:\\.(\\w+))?(?:\\.rpm)?$")
)

type Response struct {
	Srpmproc     *srpmprocpb.Cfg
	PatchName    string
	PatchContent string
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
	// srpmUrl = "https://dl.rockylinux.org/pub/rocky/8.5/BaseOS/source/tree/Packages/k/kernel-4.18.0-348.12.2.el8_5.src.rpm"
	// srpmUrl = "file:///tmp/kernelsrc/kernel-4.18.0-348.12.2.el8_5.src.rpm"
	// srpmUrl = "http://dl.rockylinux.org/pub/rocky/8.5/BaseOS/source/tree/Packages/o/openssl-1.1.1k-5.el8_5.src.rpm"

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

func fetchUpstream(upstreamRepo string, tag string, depth int32) (*git.Repository, error) {
	log.Printf("Using upstream repo: %s (tag %s)", upstreamRepo, tag)

	s := memory.NewStorage()
	fs := memfs.New()
	g, err := git.Clone(s, fs, &git.CloneOptions{
		URL:           upstreamRepo,
		ReferenceName: plumbing.NewTagReferenceName(tag),
		SingleBranch:  true,
		Depth:         int(depth),
		Progress:      os.Stdout,
		Tags:          git.NoTags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone upstream: %v", err)
	}

	return g, nil
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

func backportUpstreamChanges(distro *git.Repository, upstream *git.Repository, changes []*sidelinepb.Changes) error {
	log.Printf("Backporting upstream changes")

	distroW, err := distro.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get distro worktree: %v", err)
	}

	upstreamW, err := upstream.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get upstream worktree: %v", err)
	}

	// First do recursive overrides
	for i, change := range changes {
		log.Printf("==> Executing change %d", i)
		// Always first upstream to distro operations.
		// Currently, only "recursive_path"
		for _, recursivePath := range change.RecursivePath {
			log.Printf("\t==> Recursively copying %s", recursivePath)

			// Check that the recursive path exists in the upstream and is a directory
			fi, err := upstreamW.Filesystem.Stat(recursivePath)
			if err != nil {
				return fmt.Errorf("failed to read directory: %v", err)
			}
			if !fi.IsDir() {
				return fmt.Errorf("recursive path %s is not a directory in upstream", recursivePath)
			}

			// First delete the path from distro worktree
			err = recursivelyDeletePath(distroW.Filesystem, recursivePath)
			if err != nil {
				log.Printf("failed to remove path: %v, but continuing", err)
			}

			// Then re-create the directory
			err = distroW.Filesystem.MkdirAll(recursivePath, 0755)
			if err != nil {
				return fmt.Errorf("failed to create directory: %v", err)
			}
			_, err = distroW.Add(recursivePath)
			if err != nil {
				return fmt.Errorf("failed to add path: %v", err)
			}

			err = recursivelyCopyPath(upstreamW.Filesystem, distroW, recursivePath)
			if err != nil {
				return fmt.Errorf("failed to copy path %s: %v", recursivePath, err)
			}
		}

		// Now run file change operations
		for _, fileChange := range change.FileChange {
			log.Printf("\t==> Making changes to %s", fileChange.Path)

			// Search and replace first (and only right now)
			for _, searchReplace := range fileChange.SearchReplace {
				fileContents, err := readFileBillyFs(distroW.Filesystem, fileChange.Path)
				if err != nil {
					return fmt.Errorf("failed to read file: %v", err)
				}

				fileContents = []byte(strings.Replace(string(fileContents), searchReplace.Find, searchReplace.Replace, -1))
				err = writeFileBillyFs(distroW.Filesystem, fileChange.Path, fileContents)
				if err != nil {
					return fmt.Errorf("failed to write file: %v", err)
				}

				_, err = distroW.Add(fileChange.Path)
				if err != nil {
					return fmt.Errorf("failed to add file: %v", err)
				}
			}
		}
	}

	// show status
	status, err := distroW.Status()
	if err != nil {
		return fmt.Errorf("failed to get status: %v", err)
	}
	log.Printf("Successfully processed:\n%s", status)

	statusLines := strings.Split(status.String(), "\n")
	for _, line := range statusLines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "D") {
			path := strings.TrimPrefix(trimmed, "D ")
			_, err := distroW.Remove(path)
			if err != nil {
				return fmt.Errorf("could not delete extra file %s: %v", path, err)
			}
		}
	}

	return nil
}

func validateCfg(cfg *sidelinepb.Configuration) error {
	switch cfg.Preset {
	case "rocky8":
		cfg.PresetOverride = &sidelinepb.PresetOverride{
			Override: &sidelinepb.PresetOverride_BaseUrl{
				BaseUrl: "https://dl.rockylinux.org/pub/rocky/8",
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

	upstream, err := fetchUpstream(cfg.Upstream.GetGit(), cfg.Upstream.GetTag(), cfg.Upstream.Depth)
	if err != nil {
		return nil, err
	}

	err = backportUpstreamChanges(distro, upstream, cfg.Changes)
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
			Name:  "RESF Sideline",
			Email: "mustafa+sideline@rockylinux.org",
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
						AddToPrep: cfg.AddToPrep,
						NPath:     cfg.NPath,
					},
				},
				Changelog: []*srpmprocpb.SpecChange_ChangelogOperation{
					{
						AuthorName:  "RESF Sideline",
						AuthorEmail: "mustafa+sideline@rockylinux.org",
						Message:     []string{commitMsg},
					},
				},
			},
		},
		PatchName:    patchName,
		PatchContent: genPatch.String(),
	}

	return response, nil
}
