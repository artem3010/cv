package converter

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
)

// ConvertMarkdownToHTMLFragment конвертирует Markdown в HTML‑фрагмент без полной обёртки.
func ConvertMarkdownToHTMLFragment(mdContent string) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
		goldmark.WithRendererOptions(
			gmhtml.WithUnsafe(), // разрешает вывод сырого HTML
		),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(mdContent), &buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}
