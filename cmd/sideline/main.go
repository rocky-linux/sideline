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
	inConfig string
	output   string
)

func init() {
	root.Flags().StringVarP(&inConfig, "config", "c", "", "Config file to process")
	root.Flags().StringVarP(&output, "output", "o", ".", "Output directory")

	_ = root.MarkFlagRequired("config")
}

func mn(_ *cobra.Command, _ []string) {
	cfg, err := sideline.ParseCfg(inConfig)
	if err != nil {
		log.Fatal(err)
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
