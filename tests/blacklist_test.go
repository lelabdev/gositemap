package sitemap_test

import (
	"gositemap/sitemap"
	"os"
	"testing"
)

func TestLoadBlacklist(t *testing.T) {
	tomlContent := `exclude = ["/secret", "/admin"]`
	tmpfile, err := os.CreateTemp("", "blacklist-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(tomlContent)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	bl, err := sitemap.LoadBlacklist(tmpfile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bl.Exclude) != 2 || bl.Exclude[0] != "/secret" || bl.Exclude[1] != "/admin" {
		t.Errorf("unexpected blacklist: %+v", bl.Exclude)
	}
}
