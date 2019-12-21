package generator

import (
	"bufio"
	"bytes"
	"fmt"
	S "strings"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/styles"
	"html/template"
	"os"
	"time"
	SU "github.com/fbaube/stringutils"
	"github.com/morningconsult/serrors"
)

// WrapHtmlInMasterPageTemplate writes an index.html file.
func WrapHtmlInMasterPageTemplate(targs IndexHtmlMasterPageTemplateVariableArguments,
		blogProps SU.PropSet, dest string, t *template.Template) error {
	trunc := SU.TruncateTo(string(targs.HtmlContentFrag), 200)
	fmt.Printf("##>> WrapHtml: dest<%s> title<%s> desc<%s> tmpl?<%t> cont||%s||\n",
		dest, targs.PageTitle, targs.MetaDesc, (t!=nil), template.HTML(trunc))
	f, err := os.Create(dest)
	if err != nil {
		return serrors.Errorf("error creating file %s: %w", dest, err)
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
	blogTitle := S.TrimSuffix(blogProps["title"], ".f")
	targs.HtmlTitle = blogTitle
	if targs.PageTitle != "" {
		targs.HtmlTitle = fmt.Sprintf("%s - %s", targs.PageTitle, blogTitle)
	}
	td := new(IndexHtmlMasterPageTemplateVariables) // IndexData{
	td.IndexHtmlMasterPageTemplateVariableArguments = targs
	td.Name = blogProps["author"]
	td.Year = time.Now().Year()
	td.CanonLink = buildCanonicalLink(blogProps["url"], dest)
	td.HiliteCSS = template.CSS(hlbuf.String())

	if err := t.Execute(w, td); err != nil {
		return serrors.Errorf("error executing template %s: %w", dest, err)
	}
	if err := w.Flush(); err != nil {
		return serrors.Errorf("error writing file %s: %w", dest, err)
	}
	return nil
}
