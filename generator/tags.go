package generator

import (
	"bytes"
	"fmt"
	"html/template"
	FP "path/filepath"
	"sort"
	S "strings"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
	"github.com/morningconsult/serrors"
)

// Tag holds the data for a Tag.
type Tag struct {
	Name  string
	Link  string
	Count int
}

// ByCountDesc sorts the tags.
type ByCountDesc []*Tag

// TagsGenerator object
type TagsGenerator struct {
	Config *TagsConfig
}

// TagsConfig holds the tag's config.
type TagsConfig struct {
	TagPostsMap map[string][]*Post
	BaseConfig
}

// Generate creates the tags page.
func (g *TagsGenerator) Generate() error {
	fmt.Println("\tGenerating Tags...")
	tagPostsMap := g.Config.TagPostsMap
	t := g.Config.Template
	tagsDestDirPath := FP.Join(g.Config.Dest, "tags")
	if err := FU.ClearAndCreateDirectory(tagsDestDirPath); err != nil {
		return err
	}
	if err := generateTagIndex(tagPostsMap, t, tagsDestDirPath, g.Config.BlogProps); err != nil {
		return err
	}
	for tag, tagPosts := range tagPostsMap {
		tagPagePath := FP.Join(tagsDestDirPath, tag)
		if err := generateTagPage(tag, tagPosts, t, tagPagePath, g.Config.BlogProps); err != nil {
			return err
		}
	}
	fmt.Println("\tFinished generating Tags...")
	return nil
}

func generateTagIndex(tagPostsMap map[string][]*Post, t *template.Template,
		destDirPath string, blogProps SU.PropSet) error {
	tagsTmplPath := FP.Join("template", "tags.html")
	tagsTmplRaw, err := getTemplate(tagsTmplPath)
	if err != nil {
		return err
	}
	tags := []*Tag{}
	for tag, posts := range tagPostsMap {
		tags = append(tags, &Tag{Name: tag, Link: getTagLink(tag), Count: len(posts)})
	}
	sort.Sort(ByCountDesc(tags))
	execdTmplOutput := bytes.Buffer{}
	if err := tagsTmplRaw.Execute(&execdTmplOutput, tags); err != nil {
		return serrors.Errorf("error executing template %s: %w", tagsTmplPath, err)
	}
	// WriteIndexHTML(blogProps SU.PropSet, destDirPath, pageTitle,
	// aMetaDesc string, htmlContentFrag template.HTML, t *template.Template)
	targs := *new(IndexHtmlMasterPageTemplateVariableArguments)
	targs.PageTitle = "Tags"
	targs.HtmlTitle = "Tags"
	targs.HtmlContentFrag = template.HTML(execdTmplOutput.String())
	if err := WriteIndexHTML(targs, blogProps, destDirPath, t); err != nil {
		return err
	}
	return nil
}

func generateTagPage(tag string, posts []*Post, t *template.Template,
		destination string, blogProps SU.PropSet) error {
	if err := FU.ClearAndCreateDirectory(destination); err != nil {
		return err
	}
	pLC := new(ListingConfig)
	pLC.Posts = posts
	pLC.Template = t
	pLC.Dest = destination
	pLC.PageTitle = tag
	pLC.BlogProps =  blogProps
	println(pLC.String())
	lg := ListingGenerator{pLC}

	if err := lg.Generate(); err != nil {
		return err
	}
	return nil
}

func createTags(tagstr string) []*Tag {
	var result []*Tag
	tags := S.Split(tagstr, " ")
	for _, tag := range tags {
		result = append(result, &Tag{Name: tag, Link: getTagLink(tag)})
	}
	return result
}

func getTagLink(tag string) string {
	return fmt.Sprintf("/tags/%s/", S.ToLower(tag))
}

func (t ByCountDesc) Len() int {
	return len(t)
}

func (t ByCountDesc) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t ByCountDesc) Less(i, j int) bool {
	return t[i].Count > t[j].Count
}
