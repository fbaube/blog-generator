package generator

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// ListingData holds the data for the listing page.
type ListingData struct {
	Title      string
	Date       string
	Short      string
	Link       string
	TimeToRead string
	Tags       []*Tag
}

// ListingGenerator Object
type ListingGenerator struct {
	Config *ListingConfig
}

// ListingConfig holds the configuration for the listing page.
type ListingConfig struct {
	Posts  []*Post
	PageTitle string
	IsIndex   bool
	BaseConfig
}

func (pLC *ListingConfig) String() string {
	return fmt.Sprintf("ListCfg: %s; \n\t PgTtl<%s> IsIdx?<%t> Posts: %+v",
			pLC.BaseConfig.String(), pLC.PageTitle, pLC.IsIndex, pLC.Posts)
}

// Generate starts the listing generation.
func (g *ListingGenerator) Generate() error {
	shortTemplatePath := filepath.Join("static", "short.html")
	archiveLinkTemplatePath := filepath.Join("static", "archiveLink.html")
	posts := g.Config.Posts
	t := g.Config.Template
	destination := g.Config.Dest
	pageTitle := g.Config.PageTitle
	short, err := getTemplate(shortTemplatePath)
	if err != nil {
		return err
	}
	var postBlocks []string
	for _, post := range posts {
		meta := post.Meta
		link := fmt.Sprintf("/%s/", post.Name)
		ld := ListingData{
			Title:      meta["title"],
			Date:       meta["date"],
			Short:      meta["short"],
			Link:       link,
			Tags:       createTags(meta["tags"]),
			TimeToRead: calculateTimeToRead(string(post.HTML)),
		}
		block := bytes.Buffer{}
		if err := short.Execute(&block, ld); err != nil {
			return fmt.Errorf("error executing template %s: %v", shortTemplatePath, err)
		}
		postBlocks = append(postBlocks, block.String())
	}
	htmlBlocks := template.HTML(strings.Join(postBlocks, "<br />"))
	if g.Config.IsIndex {
		archiveLink, err := getTemplate(archiveLinkTemplatePath)
		if err != nil {
			return err
		}
		lastBlock := bytes.Buffer{}
		if err := archiveLink.Execute(&lastBlock, nil); err != nil {
			return fmt.Errorf("error executing template %s: %v", archiveLinkTemplatePath, err)
		}
		htmlBlocks = template.HTML(fmt.Sprintf("%s%s", htmlBlocks, template.HTML(lastBlock.String())))
	}
	if err := // g.Config.IndexWriter.WriteIndexHTML(destination, pageTitle, pageTitle, htmlBlocks, t); err != nil {
		WriteIndexHTML(g.Config.BlogProps, destination, pageTitle, pageTitle, htmlBlocks, t); err != nil {
		return err
	}
	return nil
}

func calculateTimeToRead(input string) string {
	// an average human reads about 200 wpm
	var secondsPerWord = 60.0 / 200.0
	// multiply with the amount of words
	words := secondsPerWord * float64(len(strings.Split(input, " ")))
	// add 12 seconds for each image
	images := 12.0 * strings.Count(input, "<img")
	result := (words + float64(images)) / 60.0
	if result < 1.0 {
		result = 1.0
	}
	return fmt.Sprintf("%.0fm", result)
}
