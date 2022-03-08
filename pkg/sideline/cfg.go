package sideline

import (
	"fmt"
	sidelinepb "github.com/rocky-linux/sideline/pb"
	"google.golang.org/protobuf/encoding/prototext"
	"io/ioutil"
	"os"
)

func ParseCfg(path string) (*sidelinepb.Configuration, error) {
	cfg := &sidelinepb.Configuration{}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	err = prototext.Unmarshal(b, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return cfg, nil
}
