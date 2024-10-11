package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Decoder decodes app's configuration from file.
//
// Supported extensions:
// - yaml, yml
// - json
type Decoder struct {
	filename string
}

// NewDecoder creates new decoder that decodes configuration file.
func NewDecoder(filename string) *Decoder {
	return &Decoder{
		filename: filename,
	}
}

// Decode configuration from file into value.
func (d *Decoder) Decode(v any) error {
	file, err := os.OpenFile(d.filename, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open file %s: %v", d.filename, err)
	}

	var dec fileDecoder

	ext := filepath.Ext(d.filename)
	switch ext {
	case ".yaml", ".yml":
		dec = yaml.NewDecoder(file)
	case ".json":
		dec = json.NewDecoder(file)
	default:
		return fmt.Errorf(
			"unsupported config file type: %s\nexpected: .yaml, .yml, .json", ext,
		)
	}

	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("decode file: %w", err)
	}

	return nil
}

type fileDecoder interface {
	Decode(interface{}) error
}
