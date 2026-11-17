package utils

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writePNG(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.White)
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeJPEG(t *testing.T, dir, name string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if err := jpeg.Encode(f, img, nil); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeText(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestFileSecurity(t *testing.T) {
	dir := t.TempDir()
	pngPath := writePNG(t, dir, "a.png")
	jpgPath := writeJPEG(t, dir, "b.jpg")
	htmlAsPng := writeText(t, dir, "c.png", "<!doctype html><html><script>alert(1)</script></html>")
	htmlPath := writeText(t, dir, "d.html", "<html>hi</html>")
	svgPath := writeText(t, dir, "e.svg", "<svg onload=alert(1)>")

	tests := []struct {
		name        string
		objectName  string
		sourceName  string
		path        string
		wantMIME    string
		wantCT      string
		wantDispo   string
	}{
		{"真实png", "GPAI5/2026-11/x.png", "原图.png", pngPath, "image/png", "image/png", "inline"},
		{"真实jpg", "GPAI5/2026-11/x.jpg", "photo.jpg", jpgPath, "image/jpeg", "image/jpeg", "inline"},
		{"伪装png", "GPAI5/2026-11/x.png", "报告.html", htmlAsPng, "", "application/octet-stream", `attachment; filename="__.html"; filename*=UTF-8''%E6%8A%A5%E5%91%8A.html`},
		{"html", "GPAI5/2026-11/x.html", "引出.html", htmlPath, "", "application/octet-stream", `attachment; filename="__.html"; filename*=UTF-8''%E5%BC%95%E5%87%BA.html`},
		{"svg", "GPAI5/2026-11/x.svg", "icon.svg", svgPath, "", "application/octet-stream", `attachment; filename="icon.svg"; filename*=UTF-8''icon.svg`},
		{"源名全失效退回随机名", "GPAI5/2026-11/F4bRiCc6BazKBccCGMJo.html", "\r\n", htmlPath, "", "application/octet-stream", `attachment; filename="F4bRiCc6BazKBccCGMJo.html"; filename*=UTF-8''F4bRiCc6BazKBccCGMJo.html`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mime := DetectSafeImageMIME(tt.objectName, tt.path)
			if mime != tt.wantMIME {
				t.Fatalf("DetectSafeImageMIME = %q, want %q", mime, tt.wantMIME)
			}
			ct, dispo := BuildOSSHeaders(mime, tt.objectName, tt.sourceName)
			if ct != tt.wantCT || dispo != tt.wantDispo {
				t.Fatalf("BuildOSSHeaders = (%q, %q), want (%q, %q)", ct, dispo, tt.wantCT, tt.wantDispo)
			}
		})
	}
}

func TestSafeDownloadName(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		want       string
	}{
		{"保留随机名和扩展名", "GPAI5/2026-11/F4bRiCc6BazKBccCGMJo.html", "F4bRiCc6BazKBccCGMJo.html"},
		{"保留多段扩展名", "GPAI5/2026-11/a.tar.gz", "a.tar.gz"},
		{"剥离引号和换行防注入", "GPAI5/2026-11/x.html\"\r\n evil", "x.htmlevil"},
		{"仅安全字符", "GPAI5/2026-11/a_b-c.HTML", "a_b-c.HTML"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := safeDownloadName(tt.objectName)
			if got != tt.want {
				t.Fatalf("safeDownloadName(%q) = %q, want %q", tt.objectName, got, tt.want)
			}
		})
	}
}
