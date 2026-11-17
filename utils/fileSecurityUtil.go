package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"
)

// 安全图片扩展名与期望 MIME 的对应关系。
// 只有「扩展名在此表中」且「文件真实内容确为该图片」时，才允许 inline 展示；
// 其余一律按 application/octet-stream + attachment 处理，避免浏览器执行 HTML/SVG/JS 等内容。
var safeImageMIME = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".webp": "image/webp",
}

// DetectSafeImageMIME 读取文件前 512 字节嗅探真实内容类型。
// 当且仅当 objectName 的扩展名在白名单内、且嗅探结果与扩展名期望一致时返回该 MIME；
// 伪装图片（如内容为 HTML 的 .png）、未知类型均返回空字符串。
func DetectSafeImageMIME(objectName, filePath string) string {
	ext := strings.ToLower(path.Ext(objectName))
	expected, ok := safeImageMIME[ext]
	if !ok {
		return ""
	}

	f, err := os.Open(filePath)
	if err != nil {
		log.Printf("检测文件类型失败: %s, %v", filePath, err)
		return ""
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return ""
	}
	if http.DetectContentType(buf[:n]) != expected {
		return ""
	}
	return expected
}

// BuildOSSHeaders 根据是否为安全图片生成 OSS 对象应答头。
// 安全图片 inline 展示（保留图床预览能力）；其余强制下载，避免浏览器执行。
//   - sourceName：用户上传时的源文件名，清洗后作为下载名（保留中文等可见字符）。
//   - objectName：OSS 对象名（系统随机名），仅当源文件名清洗后完全不可用时作为回退。
func BuildOSSHeaders(safeMIME, objectName, sourceName string) (contentType, disposition string) {
	if safeMIME != "" {
		return safeMIME, "inline"
	}
	name := sanitizeSourceName(sourceName)
	if name == "" || strings.Trim(name, ".") == "" {
		name = safeDownloadName(objectName) // 源名全失效，退回系统随机名
	}
	return "application/octet-stream", buildAttachment(name)
}

// sanitizeSourceName 清洗用户源文件名：剔除控制字符、格式字符（零宽/双向控制等隐形字符）
// 以及可能造成响应头注入或路径语义的字符，保留可见字符（含中文）。
func sanitizeSourceName(name string) string {
	// 兼容前端可能传来的 Windows 路径，只取最后一段文件名
	name = strings.ReplaceAll(name, "\\", "/")
	name = path.Base(name)
	var b strings.Builder
	for _, r := range name {
		// 控制字符（含 NUL/CR/LF）与 Unicode Cf 格式字符（零宽、方向控制、BOM 等）一律剔除
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		switch r {
		case '"', '/', '\\', ';':
			continue // 响应头注入、路径分隔相关字符
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

// buildAttachment 生成 Content-Disposition。
// filename 使用 ASCII 兼容写法（非 ASCII 用 _ 替代），filename*=UTF-8'' 使用 RFC 5987 编码，
// 两者兼顾：现代浏览器按 filename* 还原中文等原始文件名，老客户端也有可读的回退名。
func buildAttachment(name string) string {
	var ascii strings.Builder
	for _, r := range name {
		if r >= 0x20 && r <= 0x7e {
			ascii.WriteRune(r)
		} else {
			ascii.WriteByte('_')
		}
	}
	fallback := ascii.String()
	if strings.Trim(fallback, "._") == "" {
		fallback = "download"
	}
	return fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`, fallback, rfc5987Encode(name))
}

// rfc5987Encode 按 RFC 5987 对文件名做 percent 编码（保留 unreserved 字符）。
func rfc5987Encode(s string) string {
	const upperhex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(upperhex[c>>4])
			b.WriteByte(upperhex[c&0x0f])
		}
	}
	return b.String()
}

// safeDownloadName 从 OSS 对象名（系统随机名）提取回退下载文件名，
// 只保留安全字符，防止恶意文件名夹带引号/换行/控制字符造成响应头注入（CRLF injection）。
func safeDownloadName(objectName string) string {
	name := path.Base(objectName)
	var b strings.Builder
	for _, r := range name {
		// 仅允许字母、数字、点、下划线、连字符；其余（含中文、空格、引号、控制字符）一律剔除
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	name = b.String()
	if name == "" || name == "." || name == ".." {
		return "download" // 极端兜底
	}
	return name
}
