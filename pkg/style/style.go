package style

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

// Styler defines an interface to apply style to HTML content.
type Styler interface {
	Apply(content string) (string, error)
}

var baseTemplate *template.Template

func init() {
	var err error
	// Load the base template from templates/base.html
	baseTemplate, err = template.ParseFiles("templates/base.html")
	if err != nil {
		panic(fmt.Sprintf("Error loading base template: %v", err))
	}
}

// fileStyler implements the Styler interface using CSS loaded from a file.
type fileStyler struct {
	title string
	css   template.CSS // Mark CSS as safe.
}

// Apply wraps the provided content with the base template using the loaded CSS.
// The content is cast to template.HTML to prevent escaping.
func (fs *fileStyler) Apply(content string) (string, error) {
	data := struct {
		Title   string
		Style   template.CSS
		Content template.HTML
	}{
		Title:   fs.title,
		Style:   fs.css,
		Content: template.HTML(content),
	}
	var buf strings.Builder
	err := baseTemplate.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetStyler returns a Styler implementation for the given style name.
// If styleName is empty, it defaults to "Default".
// For example, if styleName == "Default", it will load the file templates/css/Default.css.
func GetStyler(styleName string) (Styler, error) {
	if strings.TrimSpace(styleName) == "" {
		styleName = "Default"
	}
	cssPath := filepath.Join("templates", "css", styleName+".css")
	data, err := os.ReadFile(cssPath)
	if err != nil {
		return nil, err
	}
	title := fmt.Sprintf("Resume — %s", styleName)
	return &fileStyler{
		title: title,
		css:   template.CSS(string(data)),
	}, nil
}

// GetAvailableStyles returns a list of style names (without extension)
// found in the templates/css directory.
func GetAvailableStyles() ([]string, error) {
	dir := filepath.Join("templates", "css")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("error reading directory %s: %w", dir, err)
	}
	var styles []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".css" {
			name := strings.TrimSuffix(entry.Name(), ".css")
			styles = append(styles, name)
		}
	}
	return styles, nil
}
