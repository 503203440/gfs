package utils

import (
	"bytes"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/bmp"
)

func TestBMPCompress(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "large.bmp")
	outPath := filepath.Join(dir, "compress_large.jpg")

	// 1. 生成 1200x600 宽度的真实 BMP 图片
	f, err := os.Create(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	srcImg := image.NewRGBA(image.Rect(0, 0, 1200, 600))
	srcImg.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if err := bmp.Encode(f, srcImg); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()

	// 2. 测试 ImageWidth 能正常读取 BMP 宽度
	width, err := ImageWidth(srcPath)
	if err != nil {
		t.Fatalf("ImageWidth 读取 BMP 失败: %v", err)
	}
	if width != 1200 {
		t.Fatalf("ImageWidth = %d, want 1200", width)
	}

	// 3. 测试 ComposeImg 缩放至 1000px 宽度（BMP 压缩后统一转 JPEG）
	if err := ComposeImg(srcPath, outPath, 1000, false); err != nil {
		t.Fatalf("ComposeImg 压缩 BMP 失败: %v", err)
	}

	// 4. 验证压缩后的图片宽度为 1000
	newWidth, err := ImageWidth(outPath)
	if err != nil {
		t.Fatalf("ImageWidth 读取压缩后的图片失败: %v", err)
	}
	if newWidth != 1000 {
		t.Fatalf("压缩后宽度 = %d, want 1000", newWidth)
	}

	// 5. 验证产物确实是 JPEG 字节（而非无压缩 BMP）
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
		t.Fatalf("压缩产物不是 JPEG 内容, 前 3 字节: %v", data[:3])
	}

	// 6. 验证安全检测能按 .jpg 扩展名识别为 image/jpeg 并 inline 展示
	mime := DetectSafeImageMIME("GPAI5_Compress/2026-11/compress_large.jpg", outPath)
	if mime != "image/jpeg" {
		t.Fatalf("DetectSafeImageMIME = %q, want image/jpeg", mime)
	}

	ct, dispo := BuildOSSHeaders(mime, "GPAI5_Compress/2026-11/compress_large.jpg", "")
	if ct != "image/jpeg" || dispo != "inline" {
		t.Fatalf("BuildOSSHeaders = (%q, %q), want (image/jpeg, inline)", ct, dispo)
	}
}
