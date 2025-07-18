package sitemap

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Blacklist struct {
	Exclude []string `toml:"exclude"`
}

func LoadBlacklist(path string) (*Blacklist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bl Blacklist
	if err := toml.Unmarshal(data, &bl); err != nil {
		return nil, err
	}
	return &bl, nil
}
