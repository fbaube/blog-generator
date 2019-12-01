package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"html/template"
	"os"
	"path/filepath"
	"time"
	SU "github.com/fbaube/stringutils"
)

// WriteIndexHTML writes an index.html file.
// func (i *IndexWriter) WriteIndexHTML(path, pageTitle, metaDescription string, content template.HTML, t *template.Template) error {
func WriteIndexHTML(blogProps SU.PropSet, path, pageTitle, metaDescription string, content template.HTML, t *template.Template) error {
	filePath := filepath.Join(path, "index.html")
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", filePath, err)
	}
	defer f.Close()
	metaDesc := metaDescription
	if metaDescription == "" {
		metaDesc = blogProps["description"]
	}
	hlbuf := bytes.Buffer{}
	hlw := bufio.NewWriter(&hlbuf)
	formatter := html.New(html.WithClasses(true))
	formatter.WriteCSS(hlw, styles.MonokaiLight)
	hlw.Flush()
	w := bufio.NewWriter(f)
	td := IndexData{
		Name:            blogProps["author"],
		Year:            time.Now().Year(),
		HTMLTitle:       getHTMLTitle(pageTitle, blogProps["title"]),
		PageTitle:       pageTitle,
		Content:         content,
		CanonicalLink:   buildCanonicalLink(path, blogProps["url"]),
		MetaDescription: metaDesc,
		HighlightCSS:    template.CSS(hlbuf.String()),
	}
	if err := t.Execute(w, td); err != nil {
		return fmt.Errorf("error executing template %s: %v", filePath, err)
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("error writing file %s: %v", filePath, err)
	}
	return nil
}
