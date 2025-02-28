package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/yuin/goldmark"
)

// generateHTML загружает нужный CSS из папки styles и конвертирует Markdown в HTML.
func generateHTML(md, styleName string) (string, error) {
	stylePath := filepath.Join("styles", styleName+".css")
	cssBytes, err := os.ReadFile(stylePath)
	if err != nil {
		return "", err
	}
	css := string(cssBytes)

	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return "", err
	}
	htmlContent := buf.String()

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
</html>`, css, htmlContent)
	return finalHTML, nil
}

const mainPageTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>CV Generator</title>
  <style>
    body { 
      font-family: Arial, sans-serif; 
      background-color: #eef2f5; 
      margin: 0; 
      padding: 0; 
    }
    header { 
      background-color: #337ab7; 
      color: white; 
      padding: 20px; 
      text-align: center; 
    }
    .page-container { 
      max-width: 1000px; 
      margin: 20px auto; 
      display: flex; 
      flex-wrap: wrap;
    }
    .editor, .preview { 
      flex: 1; 
      padding: 20px; 
      background: #fff; 
      margin: 10px; 
      border: 1px solid #ddd; 
      border-radius: 8px; 
      box-shadow: 0 2px 6px rgba(0,0,0,0.1);
    }
    #mdEditor { 
      width: 100%; 
      height: 400px; 
      border: 1px solid #ccc; 
      padding: 10px; 
      overflow-y: auto; 
      white-space: pre-wrap; 
      border-radius: 4px;
      background-color: #fafafa;
    }
    select, button { 
      margin-top: 10px; 
      padding: 8px 12px; 
      border: none; 
      border-radius: 4px; 
      background-color: #337ab7; 
      color: #fff; 
      cursor: pointer; 
    }
    button:hover { background-color: #286090; }
    /* Предпросмотр в iframe для изоляции стилей */
    #previewFrame {
      width: 100%;
      height: 400px;
      border: none;
      border-radius: 4px;
    }
    .collapsed-image {
      display: inline-block;
      max-width: 300px;
      overflow: hidden;
      text-overflow: ellipsis;
      vertical-align: bottom;
      cursor: pointer;
      border: 1px dashed #ccc;
      padding: 2px;
    }
    footer { 
      background-color: #f7f7f7; 
      text-align: center; 
      padding: 10px; 
      border-top: 1px solid #ddd; 
      margin-top: 20px;
    }
    footer a { 
      color: #337ab7; 
      text-decoration: none; 
      margin: 0 10px; 
    }
    footer a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <header>
    <h1>CV Generator</h1>
  </header>
  <div class="page-container">
    <div class="editor">
      <h2>Markdown Input</h2>
      <!-- Редактор реализован как contenteditable div -->
      <div id="mdEditor" contenteditable="true" onpaste="handlePaste(event)" placeholder="Paste your markdown here..."></div>
      <br>
      <label for="styleSelect">Select Style:</label>
      <select id="styleSelect">
        {{range .Styles}}
        <option value="{{.}}">{{.}}</option>
        {{end}}
      </select>
      <br>
      <button type="button" onclick="generateCV()">Generate Preview</button>
      <h3>Embed Image (Base64)</h3>
      <input type="file" id="imageFile" accept="image/*">
      <button type="button" onclick="embedBase64Image()">Embed Image</button>
    </div>
    <div class="preview">
      <h2>Preview</h2>
      <iframe id="previewFrame"></iframe>
      <br>
      <button onclick="downloadHTML()">Download HTML</button>
      <button onclick="downloadPDF()">Download PDF</button>
    </div>
  </div>
  <footer>
    <p>
      <a href="https://www.linkedin.com/in/aealeks3010/" target="_blank">LinkedIn</a> | 
      <a href="https://tronscan.org/#/address/TPqiD1DXrdB6j3GrVFQJ2MSz5xfk1mxucU" target="_blank">Donate (TRC20)</a>
    </p>
  </footer>
  <script>
    // Функция prepareMarkdown собирает текст редактора и заменяет все свернутые элементы на полный Markdown.
    function prepareMarkdown() {
      var editor = document.getElementById("mdEditor");
      var tempDiv = document.createElement("div");
      tempDiv.innerHTML = editor.innerHTML;
      var collapsed = tempDiv.getElementsByClassName("collapsed-image");
      var arr = Array.from(collapsed);
      arr.forEach(function(span) {
        var full = span.getAttribute("data-full");
        var textNode = document.createTextNode(full);
        span.parentNode.replaceChild(textNode, span);
      });
      return tempDiv.innerText;
    }
    // Отправка Markdown и выбранного стиля для генерации предпросмотра.
    function generateCV(){
      var md = prepareMarkdown();
      var style = document.getElementById("styleSelect").value;
      var xhr = new XMLHttpRequest();
      xhr.open("POST", "/generate", true);
      xhr.setRequestHeader("Content-Type", "application/x-www-form-urlencoded");
      xhr.onreadystatechange = function(){
        if(xhr.readyState === 4 && xhr.status === 200){
          document.getElementById("previewFrame").srcdoc = xhr.responseText;
        }
      };
      xhr.send("md=" + encodeURIComponent(md) + "&style=" + encodeURIComponent(style));
    }
    // Встраивание изображения (через выбор файла) – конвертация в Base64.
    function embedBase64Image(){
      var fileInput = document.getElementById("imageFile");
      if(fileInput.files.length === 0){
        alert("Please select an image file first.");
        return;
      }
      var file = fileInput.files[0];
      var reader = new FileReader();
      reader.onload = function(e){
        var base64Data = e.target.result;
        var fullMarkdown = "![Embedded Image](" + base64Data + ")";
        var truncated = "![Embedded Image](" + base64Data.substring(0,30) + "...)";
        var span = "<span class='collapsed-image' data-full='" + fullMarkdown.replace(/'/g, "&#39;") + "' onclick='expandImage(this)'>" + truncated + "</span>";
        document.getElementById("mdEditor").innerHTML += "\n\n" + span + "\n\n";
      };
      reader.readAsDataURL(file);
    }
    // Обработка события paste (Ctrl+V) для вставки изображения.
    function handlePaste(e){
      var clipboardData = e.clipboardData || window.clipboardData;
      if(!clipboardData) return;
      var items = clipboardData.items;
      if(!items) return;
      for(var i=0; i<items.length; i++){
        var item = items[i];
        if(item.kind === 'file' && item.type.indexOf('image/') === 0){
          e.preventDefault();
          var file = item.getAsFile();
          embedImageFromPaste(file);
        }
      }
    }
    function embedImageFromPaste(file){
      var reader = new FileReader();
      reader.onload = function(e){
        var base64Data = e.target.result;
        var fullMarkdown = "![Pasted Image](" + base64Data + ")";
        var truncated = "![Pasted Image](" + base64Data.substring(0,30) + "...)";
        var span = "<span class='collapsed-image' data-full='" + fullMarkdown.replace(/'/g, "&#39;") + "' onclick='expandImage(this)'>" + truncated + "</span>";
        document.getElementById("mdEditor").innerHTML += "\n\n" + span + "\n\n";
      };
      reader.readAsDataURL(file);
    }
    // Разворачиваем свернутый элемент изображения.
    function expandImage(spanElem){
      spanElem.outerHTML = spanElem.getAttribute("data-full");
    }
    // Скачать HTML.
    function downloadHTML(){
      var md = prepareMarkdown();
      var style = document.getElementById("styleSelect").value;
      var form = document.createElement("form");
      form.method = "POST";
      form.action = "/download/html";
      var mdInput = document.createElement("input");
      mdInput.type = "hidden";
      mdInput.name = "md";
      mdInput.value = md;
      form.appendChild(mdInput);
      var styleInput = document.createElement("input");
      styleInput.type = "hidden";
      styleInput.name = "style";
      styleInput.value = style;
      form.appendChild(styleInput);
      document.body.appendChild(form);
      form.submit();
    }
    // Скачать PDF.
    function downloadPDF(){
      var md = prepareMarkdown();
      var style = document.getElementById("styleSelect").value;
      var form = document.createElement("form");
      form.method = "POST";
      form.action = "/download/pdf";
      var mdInput = document.createElement("input");
      mdInput.type = "hidden";
      mdInput.name = "md";
      mdInput.value = md;
      form.appendChild(mdInput);
      var styleInput = document.createElement("input");
      styleInput.type = "hidden";
      styleInput.name = "style";
      styleInput.value = style;
      form.appendChild(styleInput);
      document.body.appendChild(form);
      form.submit();
    }
  </script>
</body>
</html>
`

// mainPageHandler отдает главную страницу.
func mainPageHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("styles")
	if err != nil {
		http.Error(w, "Cannot read styles directory", http.StatusInternalServerError)
		return
	}
	var styles []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".css" {
			styles = append(styles, strings.TrimSuffix(f.Name(), ".css"))
		}
	}
	tmpl, err := template.New("main").Parse(mainPageTemplate)
	if err != nil {
		http.Error(w, "Template parse error", http.StatusInternalServerError)
		return
	}
	data := struct{ Styles []string }{Styles: styles}
	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Template execute error", http.StatusInternalServerError)
		return
	}
}

func generateHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	md := r.FormValue("md")
	style := r.FormValue("style")
	finalHTML, err := generateHTML(md, style)
	if err != nil {
		http.Error(w, "Error generating HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, finalHTML)
}

func downloadHTMLHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	md := r.FormValue("md")
	style := r.FormValue("style")
	finalHTML, err := generateHTML(md, style)
	if err != nil {
		http.Error(w, "Error generating HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=cv.html")
	io.WriteString(w, finalHTML)
}

func downloadPDFHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}
	md := r.FormValue("md")
	style := r.FormValue("style")
	finalHTML, err := generateHTML(md, style)
	if err != nil {
		http.Error(w, "Error generating PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}
	pdfBytes, err := generatePDFWithChrome(finalHTML)
	if err != nil {
		http.Error(w, "Error generating PDF: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=cv.pdf")
	w.Write(pdfBytes)
}

// generatePDFWithChrome использует chromedp для конвертации HTML в PDF.
func generatePDFWithChrome(htmlContent string) ([]byte, error) {
	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	dataURL := "data:text/html," + url.PathEscape(htmlContent)
	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate(dataURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, err
	}
	return pdfBuf, nil
}

func main() {
	http.HandleFunc("/", mainPageHandler)
	http.HandleFunc("/generate", generateHandler)
	http.HandleFunc("/download/html", downloadHTMLHandler)
	http.HandleFunc("/download/pdf", downloadPDFHandler)

	port := flag.String("port", "8080", "Port to listen on")
	flag.Parse()
	log.Printf("Server running on http://localhost:%s\n", *port)
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
