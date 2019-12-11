package generator

import(
  	"html/template"
)

// IndexHtmlMasterPageTemplateVariableArguments
// is variables passed to func WriteIndexHTML(..)
type IndexHtmlMasterPageTemplateVariableArguments struct {
  HtmlTitle   string
  PageTitle   string
  MetaDesc    string
  HtmlContentFrag template.HTML
}

// IndexHtmlMasterPageTemplateVariables used to be
// IndexData: a data container for the landing page.
type IndexHtmlMasterPageTemplateVariables struct {
  IndexHtmlMasterPageTemplateVariableArguments
  CanonLink string
  HiliteCSS template.CSS
	Name      string
	Year      int
}
