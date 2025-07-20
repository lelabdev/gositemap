package sitemap_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func buildBinary(t *testing.T, tmpdir string) string {
	binRoot := "gositemap-test-bin"
	cmd := exec.Command("go", "build", "-o", binRoot, "main.go", "cli.go")
	cmd.Dir = ".." // project root
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	binSrc := filepath.Join("..", binRoot)
	binTmp := filepath.Join(tmpdir, "gositemap-test-bin")
	in, err := os.Open(binSrc)
	if err != nil {
		t.Fatalf("failed to open built binary: %v", err)
	}
	defer in.Close()
	outf, err := os.Create(binTmp)
	if err != nil {
		t.Fatalf("failed to create binary in tmpdir: %v", err)
	}
	defer outf.Close()
	if _, err := io.Copy(outf, in); err != nil {
		t.Fatalf("failed to copy binary: %v", err)
	}
	os.Remove(binSrc)
	if err := os.Chmod(binTmp, 0755); err != nil {
		t.Fatalf("failed to chmod binary: %v", err)
	}
	return binTmp
}

func TestHelpOption(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	cmd := exec.Command(bin, "--help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help should exit 0, got %v", err)
	}
	if !strings.Contains(string(out), "GoSitemap") || !strings.Contains(string(out), "gositemap.toml") {
		t.Errorf("help output missing expected content: %s", out)
	}
}

func TestQuietOption(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	// Write minimal config
	blogContentDir := filepath.Join(tmp, "src", "lib", "content")
	os.MkdirAll(blogContentDir, 0755)
	os.WriteFile(filepath.Join(blogContentDir, "foo.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "gositemap.toml"), []byte(fmt.Sprintf("base_url = \"https://mysite.com\"\n[content_types]\nblog = \"%s\"\n", strings.ReplaceAll(blogContentDir, "\\", "/"))), 0644)
	cmd := exec.Command(bin, "--quiet", "--dry-run")
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("quiet+dry-run should exit 0, got %v", err)
	}
	if strings.Contains(string(out), "Detected") || strings.Contains(string(out), "Sitemap successfully generated") {
		t.Errorf("quiet should suppress logs, got: %s", out)
	}
	if !strings.Contains(string(out), "<urlset") {
		t.Errorf("sitemap xml missing in dry-run: %s", out)
	}
}

func TestInvalidBaseURL(t *testing.T) {
	tmp := t.TempDir()
	bin := buildBinary(t, tmp)
	blogContentDir := filepath.Join(tmp, "src", "lib", "content")
	os.MkdirAll(blogContentDir, 0755)
	os.WriteFile(filepath.Join(blogContentDir, "foo.md"), []byte(""), 0644)
	os.WriteFile(filepath.Join(tmp, "gositemap.toml"), []byte(fmt.Sprintf("base_url = \"not-a-url\"\n[content_types]\nblog = \"%s\"\n", strings.ReplaceAll(blogContentDir, "\\", "/"))), 0644)
	cmd := exec.Command(bin)
	cmd.Dir = tmp
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("should fail with invalid base_url")
	}
	if !strings.Contains(string(out), "Invalid base_url") {
		t.Errorf("missing error for invalid base_url: %s", out)
	}
}
