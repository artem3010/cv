package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
)

func main() {
	// Флаг для входного Markdown-файла, по умолчанию cv.md
	inputFile := flag.String("file", "cv.md", "Path to the markdown file (e.g. cv.md)")

	// Флаг для выбора стиля (без расширения). По умолчанию "default"
	styleName := flag.String("style", "default", "Name of the style to apply (without .css extension)")

	flag.Parse()

	// Считываем содержимое markdown-файла
	mdContent, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %v", *inputFile, err)
	}

	// Конвертируем Markdown → HTML с помощью goldmark
	var buf bytes.Buffer
	if err := goldmark.Convert(mdContent, &buf); err != nil {
		log.Fatalf("Error converting markdown to HTML: %v", err)
	}
	htmlContent := buf.String()

	// Формируем путь к файлу стилей, например "styles/default.css"
	styleFilePath := filepath.Join("styles", *styleName+".css")

	// Читаем CSS из файла
	styleContent, err := os.ReadFile(styleFilePath)
	if err != nil {
		log.Fatalf("Error reading style file %s: %v", styleFilePath, err)
	}

	// Встраиваем CSS и полученный HTML-контент в итоговый HTML-документ
	finalHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>CV</title>
	<style>
	%s
	</style>
</head>
<body>
	<div class="container">
%s
	</div>
</body>
</html>`, styleContent, htmlContent)

	// Определяем имя выходного файла (меняем .md → .html)
	outputFile := strings.TrimSuffix(*inputFile, ".md") + ".html"
	if err := os.WriteFile(outputFile, []byte(finalHTML), 0644); err != nil {
		log.Fatalf("Error writing HTML file: %v", err)
	}

	fmt.Printf("HTML CV successfully generated: %s (using style: %s)\n", outputFile, *styleName)
}
