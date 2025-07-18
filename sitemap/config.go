package sitemap

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	BaseURL      string            `toml:"base_url"`
	Exclude      []string          `toml:"exclude"`
	ContentTypes map[string]string `toml:"content_types"`
	ChangeFreq   map[string]string `toml:"changefreq"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
