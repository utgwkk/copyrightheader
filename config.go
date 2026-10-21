package copyrightheader

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// FileConfig is the YAML config file structure.
type FileConfig struct {
	Header        string   `yaml:"header"`
	PatternStart  string   `yaml:"pattern_start"`
	PatternEnd    string   `yaml:"pattern_end"`
	PatternIgnore string   `yaml:"pattern_ignore"`
	Exclude       []string `yaml:"exclude"`
}

// LoadConfig reads a YAML config file and returns a Config.
func LoadConfig(path string) (Config, error) {
	// #nosec G304: path is the user-supplied -config flag value; reading it is the intended behavior
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading %s: %w", path, err)
	}
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", path, err)
	}
	return Config{
		Header:        fc.Header,
		PatternStart:  fc.PatternStart,
		PatternEnd:    fc.PatternEnd,
		PatternIgnore: fc.PatternIgnore,
		Exclude:       fc.Exclude,
	}, nil
}
