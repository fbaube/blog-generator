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
	"github.com/morningconsult/serrors"
)

// WriteIndexHTML writes an index.html file.
func WriteIndexHTML(targs IndexHtmlMasterPageTemplateVariableArguments,
	blogProps SU.PropSet, destDirPath string, t *template.Template) error {
	fmt.Printf("##>> WrtIdxHtml: dest<%s> title<%s> desc<%s> tmpl?<%t> cont||%s||\n",
		destDirPath, targs.PageTitle, targs.MetaDesc, (t!=nil), targs.HtmlContentFrag)
	filePath := filepath.Join(destDirPath, "index.html")
	f, err := os.Create(filePath)
	if err != nil {
		return serrors.Errorf("error creating file %s: %w", filePath, err)
	}
	defer f.Close()
	if targs.MetaDesc == "" {
		targs.MetaDesc = blogProps["description"]
	}
	hlbuf := bytes.Buffer{}
	hlw := bufio.NewWriter(&hlbuf)
	formatter := html.New(html.WithClasses(true))
	formatter.WriteCSS(hlw, styles.MonokaiLight)
	hlw.Flush()
	w := bufio.NewWriter(f)

	// var blogTitle, htmlTitle string
	blogTitle := blogProps["title"]
	targs.HtmlTitle = blogTitle
	if targs.PageTitle != "" {
		targs.HtmlTitle = fmt.Sprintf("%s - %s", targs.PageTitle, blogTitle)
	}
	td := new(IndexHtmlMasterPageTemplateVariables) // IndexData{
	td.IndexHtmlMasterPageTemplateVariableArguments = targs
	td.Name = blogProps["author"]
	td.Year = time.Now().Year()
	td.CanonLink = buildCanonicalLink(blogProps["url"], destDirPath)
	td.HiliteCSS = template.CSS(hlbuf.String())

	if err := t.Execute(w, td); err != nil {
		return serrors.Errorf("error executing template %s: %w", filePath, err)
	}
	if err := w.Flush(); err != nil {
		return serrors.Errorf("error writing file %s: %w", filePath, err)
	}
	return nil
}
