package handler

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"text/template"

	"github.com/artem3010/cv/pkg/converter"
	"github.com/artem3010/cv/pkg/style"
)

const FormTemplatePath = "templates/form.html"

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tmpl, err := template.ParseFiles(FormTemplatePath)
		if err != nil {
			http.Error(w, "Error during form template loading", http.StatusInternalServerError)
			return
		}
		styles, err := style.GetAvailableStyles()
		if err != nil {
			http.Error(w, "Error during styles loading: "+err.Error(), http.StatusInternalServerError)
			return
		}
		data := struct {
			Styles []string
		}{
			Styles: styles,
		}
		tmpl.Execute(w, data)

	case http.MethodPost:
		username := r.FormValue("username")
		branch := r.FormValue("branch")
		styleName := r.FormValue("style")
		action := r.FormValue("action")
		scaleStr := r.FormValue("scale")

		if username == "" {
			http.Error(w, "Here's no username", http.StatusBadRequest)
			return
		}
		if branch == "" {
			branch = "main"
		}

		scaleFloat, err := strconv.ParseFloat(scaleStr, 64)
		if err != nil {
			scaleFloat = 100 // если не получилось считать, ставим 100%
		}

		mdContent, err := fetchReadme(username, branch)
		if err != nil {
			if branch == "main" {
				branch = "master"
			} else {
				branch = "main"
			}
			mdContent, err = fetchReadme(username, branch)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Couldn't get README.md for user " + username))
				return
			}
		}

		fragment, err := converter.ConvertMarkdownToHTMLFragment(mdContent)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't get README.md for user " + username))
			return
		}

		styler, err := style.GetStyler(styleName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't get a style: " + err.Error()))
			return
		}

		htmlContent, err := styler.Apply(fragment)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Couldn't get a style: " + err.Error()))
			return
		}

		if action == "html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, htmlContent)
		} else if action == "pdf" {
			// Передаём масштаб в PDF (делим на 100, т.к. WithScale(1.0) = 100%)
			pdfBytes, err := converter.ConvertHTMLToPDF(htmlContent, scaleFloat/100.0)
			if err != nil {
				http.Error(w, "Couldn't convert to PDF: "+err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", "attachment; filename=resume.pdf")
			w.Write(pdfBytes)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Wrong action"))
		}

	default:
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Method is not implemented"))
	}
}

func fetchReadme(username, branch string) (string, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/README.md", username, username, branch)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP статус %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
