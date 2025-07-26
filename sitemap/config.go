package sitemap

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Glob struct {
	Paths []string `toml:"paths"`
}

type Config struct {
	BaseURL      string            `toml:"base_url"`
	OutputPath   string            `toml:"output_path"`
	PreserveExisting *bool             `toml:"preserve_existing"`
	ContentTypes map[string]string `toml:"content_types"`
	ChangeFreq   map[string]string `toml:"changefreq"`
	Exclude      []string          `toml:"exclude"`
	Glob         []Glob            `toml:"glob"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Debug print to see the content being unmarshaled
	// fmt.Printf("Debug: Loading config from %s\nContent:\n%s\n", path, string(data))
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
