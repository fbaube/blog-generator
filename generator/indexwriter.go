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

// IndexWriter writes index.html files.
type IndexWriter struct {
	BlogTitle  string
	BlogDesc   string
	BlogAuthor string
	BlogURL    string
}

func NewIndexWriter(cfg []SU.PropSet) (*IndexWriter) {
	iw := new(IndexWriter)
	iw.BlogTitle  = cfg[1]["title"]
	iw.BlogDesc   = cfg[1]["description"]
	iw.BlogAuthor = cfg[1]["author"]
	iw.BlogURL    = cfg[1]["url"]
	return iw
}

// WriteIndexHTML writes an index.html file.
func (i *IndexWriter) WriteIndexHTML(path, pageTitle, metaDescription string, content template.HTML, t *template.Template) error {
	filePath := filepath.Join(path, "index.html")
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", filePath, err)
	}
	defer f.Close()
	metaDesc := metaDescription
	if metaDescription == "" {
		metaDesc = i.BlogDesc
	}
	hlbuf := bytes.Buffer{}
	hlw := bufio.NewWriter(&hlbuf)
	formatter := html.New(html.WithClasses(true))
	formatter.WriteCSS(hlw, styles.MonokaiLight)
	hlw.Flush()
	w := bufio.NewWriter(f)
	td := IndexData{
		Name:            i.BlogAuthor,
		Year:            time.Now().Year(),
		HTMLTitle:       getHTMLTitle(pageTitle, i.BlogTitle),
		PageTitle:       pageTitle,
		Content:         content,
		CanonicalLink:   buildCanonicalLink(path, i.BlogURL),
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
