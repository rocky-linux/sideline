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

package main

import (
	"github.com/rocky-linux/sideline/pkg/sideline"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/prototext"
	"io/ioutil"
	"log"
	"path/filepath"
)

var root = &cobra.Command{
	Use:   "sideline",
	Short: "sideline is a tool for managing backports into distribution packages from the upstream source",
	Run:   mn,
}

var (
	inConfig     string
	output       string
	defaultEmail string
	defaultName  string
)

func init() {
	root.Flags().StringVarP(&inConfig, "config", "c", "", "Config file to process")
	root.Flags().StringVarP(&output, "output", "o", ".", "Output directory")
	root.Flags().StringVarP(&defaultEmail, "default-email", "e", "releng+sideline@rockylinux.org", "Default email address")
	root.Flags().StringVarP(&defaultName, "default-name", "n", "RESF Sideline (Backporter)", "Default name")

	_ = root.MarkFlagRequired("config")
}

func mn(_ *cobra.Command, _ []string) {
	cfg, err := sideline.ParseCfg(inConfig)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.ContactEmail == "" {
		cfg.ContactEmail = defaultEmail
	}
	if cfg.ContactName == "" {
		cfg.ContactName = defaultName
	}

	resp, err := sideline.Run(cfg)
	if err != nil {
		log.Fatal(err)
	}

	directivesPath := filepath.Join(output, "directives.cfg")
	marshalOptions := prototext.MarshalOptions{
		Indent:    "  ",
		Multiline: true,
	}
	directives, err := marshalOptions.Marshal(resp.Srpmproc)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(directivesPath, directives, 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote file %s", directivesPath)

	patchPath := filepath.Join(output, resp.PatchName)
	err = ioutil.WriteFile(patchPath, []byte(resp.PatchContent), 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote patch %s", patchPath)
}

func main() {
	if err := root.Execute(); err != nil {
		log.Fatalf("could not execute root command: %v", err)
	}
}
