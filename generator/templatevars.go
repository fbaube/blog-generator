package generator

import(
  	"html/template"
)

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

/*
// IndexData is a data container for the landing page.
type IndexData struct {
  IndexHtmlMasterPageTemplateVariableArguments
	Year            int
	Name            string
	CanonicalLink   string
	HighlightCSS    template.CSS
}
*/
