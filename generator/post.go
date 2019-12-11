package generator

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/alecthomas/chroma/formatters/html"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/yuin/goldmark"
	"html/template"
	"io/ioutil"
	"os"
	FP "path/filepath"
	"strings"
	"time"
	"errors"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
	"github.com/morningconsult/serrors"
)

// Post holds data for a post.
type Post struct {
	// Name      string
	Dir      *FU.CheckedPath
	DirBase	  string
	File     *FU.CheckedPath
	SU.PropSet
	ContentMD string
	CntAsHTML string
	ImagesDir string
	Images    []string
	// YAML "ParsedDate" is taken from the PropSet
	// as a string and then parsed into this field
	ParsedDate time.Time
}

// ByDateDesc is the sorting object for posts.
type ByDateDesc []*Post
//  dateFormat is for sorting posts
var dateFormat string

// PostGenerator object
type PostGenerator struct {
	Config *PostConfig
}

// PostConfig holds the post's configuration.
type PostConfig struct {
	Post        *Post
	BaseConfig
}

func (pPC *PostConfig) String() string {
	return fmt.Sprintf("PostCfg: %s; \n\t Post: %+v",
			pPC.BaseConfig.String(), *pPC.Post)
}

// Generate generates a post.
func (g *PostGenerator) Generate() error {
	post := g.Config.Post
	destination := g.Config.Dest
	t := g.Config.Template
	fmt.Printf("\tGenerating Post: %s...\n", post.PropSet["title"])
	staticDirPath := FP.Join(destination, post.DirBase)
	if err := os.Mkdir(staticDirPath, os.ModePerm); err != nil {
		return serrors.Errorf("error creating directory at %s: %w", staticDirPath, err)
	}
	if post.ImagesDir != "" {
		if err := copyImagesDir(post.ImagesDir, staticDirPath); err != nil {
			return err
		}
	}
	// func WriteIndexHTML(blogProps SU.PropSet, destDirPath, pageTitle,
	// aMetaDesc string, htmlContentFrag template.HTML, t *template.Template)
	targs := *new(IndexHtmlMasterPageTemplateVariableArguments)
	targs.PageTitle = post.PropSet["title"]
	targs.HtmlTitle = post.PropSet["title"]
	targs.MetaDesc = post.PropSet["short"]
	targs.HtmlContentFrag = template.HTML(post.CntAsHTML)
	if err := WriteIndexHTML(targs, g.Config.BlogProps, staticDirPath, t); err != nil {
		return err
	}
	fmt.Printf("\tFinished generating Post: %s...\n", post.PropSet["title"])
	return nil
}

func newPost(dirpath, dateFormat string) (p *Post, e error) {
	println("newPost dir:", dirpath)
	p, e = getPost(dirpath)
	if e != nil {
		return nil, e
	}
	// println("newPost:", dirpath, "||||", p.CntAsHTML, "||||")
	p.ImagesDir, p.Images, e = getImages(dirpath)
	if e != nil {
		return nil, e
	}
	p.DirBase = FP.Base(p.Dir.String())
	return p, nil
	// &Post{Name: name, PropSet: postYamlMeta, CntAsHTML: string(postHtml), ImagesDir: imagesDir, Images: images}, nil
}

func copyImagesDir(source, destination string) (err error) {
	path := FP.Join(destination, "images")
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return serrors.Errorf("error creating images directory at %s: %w", path, err)
	}
	files, err := ioutil.ReadDir(source)
	if err != nil {
		return serrors.Errorf("error reading directory %s: %w", path, err)
	}
	for _, file := range files {
		src := FP.Join(source, file.Name())
		dst := FP.Join(path, file.Name())
		if err := FU.CopyFileFromTo(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// getPost reads "post.md" and returns the YAML header metadata, the
// post's content Markdown, and the post's content converted to HTML.
func getPost(path string) (p *Post, e error) {
	p = new(Post)
	p.Dir = FU.NewCheckedPath(path)
	if !(p.Dir.Exists && p.Dir.IsDir) {
		return nil, errors.New("Not a readable directory: " + path)
	}
	postFP := FP.Join(path, "post.md")
	p.File = FU.NewCheckedPath(postFP)
	if !(p.File.Exists && p.File.IsFile) {
		return nil, errors.New("Not a readable file: " + postFP)
	}
	p.File = p.File.LoadFile()
	if e = p.File.GetError(); e != nil {
		return nil, serrors.Errorf("Can't load file <%s>: %w", postFP, e)
	}
	// println("RAW:", p.File.Raw)
	// Try to evaluate YAML metadata header
	// func GetYamlMetadata(instr string) (map[string]interface{}, string, error) {
	p.PropSet, p.ContentMD, e = SU.GetYamlMetadataAsPropSet(p.File.Raw)
	if e != nil {
		panic("post.go: error from Yaml")
	}
	fmt.Printf("##>> getPost: PropSet: %+v \n", p.PropSet)
	// Replace BlackFriday with Goldmark
	// html := blackfriday.MarkdownCommon(input)
	var bb bytes.Buffer
	if e = goldmark.Convert([]byte(p.ContentMD), &bb); e != nil {
  	panic(e)
	}
	p.CntAsHTML, e = replaceCodeParts([]byte(bb.String()))
	if e != nil {
		return nil, serrors.Errorf("error during syntax hiliting of %s: %w", postFP, e)
	}
	return p, nil
}

func getImages(path string) (string, []string, error) {
	dirPath := FP.Join(path, "images")
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, nil
		}
		return "", nil, serrors.Errorf("error while reading folder %s: %w", dirPath, err)
	}
	images := []string{}
	for _, file := range files {
		images = append(images, file.Name())
	}
	return dirPath, images, nil
}

func replaceCodeParts(htmlFile []byte) (string, error) {
	byteReader := bytes.NewReader(htmlFile)
	doc, err := goquery.NewDocumentFromReader(byteReader)
	if err != nil {
		return "", serrors.Errorf("error while parsing html: %w", err)
	}
	// find code-parts via css selector and replace them with highlighted versions
	doc.Find("code[class*=\"language-\"]").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		lang := strings.TrimPrefix(class, "language-")
		oldCode := s.Text()
		lexer := lexers.Get(lang)
		formatter := html.New(html.WithClasses(true))
		iterator, err := lexer.Tokenise(nil, string(oldCode))
		if err != nil {
			fmt.Printf("ERROR during syntax highlighting, %v", err)
		}
		b := bytes.Buffer{}
		buf := bufio.NewWriter(&b)
		err = formatter.Format(buf, styles.GitHub, iterator)
		if err != nil {
			fmt.Printf("ERROR during syntax highlighting, %v", err)
		}
		buf.Flush()
		s.SetHtml(b.String())
	})
	new, err := doc.Html()
	if err != nil {
		return "", serrors.Errorf("error while generating html: %w", err)
	}
	// replace unnecessarily added html tags
	new = strings.Replace(new, "<html><head></head><body>", "", 1)
	new = strings.Replace(new, "</body></html>", "", 1)
	return new, nil
}

func (p ByDateDesc) Len() int {
	return len(p)
}

func (p ByDateDesc) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p ByDateDesc) Less(i, j int) bool {
	return p[i].ParsedDate.After(p[j].ParsedDate)
}
