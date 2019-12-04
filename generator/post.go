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
	// "gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
)

// Post holds data for a post.
type Post struct {
	Name      string
	TheDir    FU.CheckedPath
	ThePost   FU.CheckedPath
	Meta      SU.PropSet
	HTML      string
	ImagesDir string
	Images    []string
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
			pPC.BaseConfig.String(), pPC.Post)
}

// Generate generates a post.
func (g *PostGenerator) Generate() error {
	post := g.Config.Post
	destination := g.Config.Dest
	t := g.Config.Template
	fmt.Printf("\tGenerating Post: %s...\n", post.Meta["title"])
	staticPath := filepath.Join(destination, post.Name)
	if err := os.Mkdir(staticPath, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory at %s: %v", staticPath, err)
	}
	if post.ImagesDir != "" {
		if err := copyImagesDir(post.ImagesDir, staticPath); err != nil {
			return err
		}
	}
	if err := WriteIndexHTML(g.Config.BlogProps, staticPath, post.Meta["title"],
			post.Meta["short"], template.HTML(string(post.HTML)), t); err != nil {
		return err
	}
	fmt.Printf("\tFinished generating Post: %s...\n", post.Meta["title"])
	return nil
}

func newPost(path, dateFormat string) (*Post, error) {
	println("newPost:", path)

	// Can ignore the post Markdown
	postYamlMeta, _, postHtml, err := getPost(path)
	if err != nil {
		return nil, err
	}
	println("newPost:", path, "||||", postHtml, "||||")
	imagesDir, images, err := getImages(path)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)

	return &Post{Name: name, Meta: postYamlMeta, HTML: string(postHtml), ImagesDir: imagesDir, Images: images}, nil
}

func copyImagesDir(source, destination string) (err error) {
	path := filepath.Join(destination, "images")
	if err := os.Mkdir(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating images directory at %s: %v", path, err)
	}
	files, err := ioutil.ReadDir(source)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", path, err)
	}
	for _, file := range files {
		src := filepath.Join(source, file.Name())
		dst := filepath.Join(path, file.Name())
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

/*
func getMeta(path, dateFormat string) (SU.PropSet, error) {
	filePath := filepath.Join(path, "meta.yml")
	metaraw, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error while reading file %s: %v", filePath, err)
	}
	meta := Meta{}
	err = yaml.Unmarshal(metaraw, &meta)
	if err != nil {
		return nil, fmt.Errorf("error reading yml in %s: %v", filePath, err)
	}
	parsedDate, err := time.Parse(dateFormat, meta.Date)
	if err != nil {
		return nil, fmt.Errorf("error parsing date in %s: %v", filePath, err)
	}
	meta.ParsedDate = parsedDate
	return &meta, nil
}
*/

// getPost reads "post.md" and returns the YAML header metadata, the
// post's content Markdown, and the post's content converted to HTML.
func getPost(path string) (hdrMeta SU.PropSet, cntMD, cntAsHtml string, err error) {
	filePath := filepath.Join(path, "post.md")
	input, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, "", "", fmt.Errorf("error reading file <%s>: %w", filePath, err)
	}
	// Try to evaluate YAML metadata header
	// func GetYamlMetadata(instr string) (map[string]interface{}, string, error) {
	psMeta, mdCont, e := SU.GetYamlMetadataAsPropSet(string(input))
	if psMeta == nil {
		panic("post.go: nil from Yaml")
	}
	if e != nil {
		panic("post.go: error from Yaml")
	}
	fmt.Printf("##>> getPost: PropSet: %+v \n", psMeta)
	// Replace BlackFriday with Goldmark
	// html := blackfriday.MarkdownCommon(input)
	var htmlCont bytes.Buffer
	if err := goldmark.Convert([]byte(mdCont), &htmlCont); err != nil {
  	panic(err)
	}
  html2 := []byte(htmlCont.String())
	replaced, err := replaceCodeParts(html2)
	if err != nil {
		return nil, "", "", fmt.Errorf("error during syntax hiliting of %s: %v", filePath, err)
	}
	return psMeta, mdCont, replaced, nil
}

func getImages(path string) (string, []string, error) {
	dirPath := filepath.Join(path, "images")
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, nil
		}
		return "", nil, fmt.Errorf("error while reading folder %s: %v", dirPath, err)
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
		return "", fmt.Errorf("error while parsing html: %v", err)
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
		return "", fmt.Errorf("error while generating html: %v", err)
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
