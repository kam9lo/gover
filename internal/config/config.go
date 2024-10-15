package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Option is an option for selection prompt where value fills provided in
// template field and version specifies its impact on semver version.
type Option struct {
	Value       string `json:"value" yaml:"value" validate:"required"`
	Description string `json:"description" yaml:"description" validate:""`
	Version     string `json:"version" yaml:"version" validate:"oneof=major minor patch"`
}

// Field implements selectable option field.
func (o Option) Field() string {
	return o.Value
}

// Field implements selectable option field documentation.
func (o Option) Doc() string {
	return o.Description
}

// Config contains structure of configuration file.
type Config struct {
	Templates struct {
		Commit    string `json:"commit" yaml:"commit" validate:"required"`
		Changelog string `json:"changelog" yaml:"changelog" validate:"-"`
	} `json:"templates" yaml:"templates"`
	Args []struct {
		Name     string   `json:"name,omitempty" yaml:"name" validate:"required"`
		Options  []Option `json:"options,omitempty" yaml:"options" validate:"-"`
		Required bool     `json:"required,omitempty" yaml:"required,omitempty" validate:"-"`
		Width    int      `json:"width,omitempty" yaml:"width" validate:"omitempty,gte=0"`
	} `json:"args,omitempty" yaml:"args" validate:"gt=0"`
}

func (c *Config) RequiredArgs() (r []string) {
	for _, arg := range c.Args {
		if arg.Required {
			r = append(r, arg.Name)
		}
	}
	return
}

// NewFromFile returns configuration from file with given path.
func NewFromFile(path string) (*Config, error) {
	cfg := &Config{}

	if err := NewDecoder(path).Decode(cfg); err != nil {
		return nil, fmt.Errorf("new decoder decode: %w", err)
	}
	if err := validator.New().Struct(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}
