package generator

import (
	"bytes"
	"fmt"
	"html/template"
	FP "path/filepath"
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
	return fmt.Sprintf("LstgCfg: %s; \n\t PgTtl<%s> IsIdx?<%t> Posts: %+v",
			pLC.BaseConfig.String(), pLC.PageTitle, pLC.IsIndex, pLC.Posts)
}

// Generate starts the listing generation.
func (g *ListingGenerator) Generate() error {
	shortTemplatePath := FP.Join("template", "short.html")
	archiveLinkTemplatePath := FP.Join("template", "archiveLink.html")
	posts := g.Config.Posts
	t := g.Config.Template
	destDirPath := g.Config.Dest
	pageTitle := g.Config.PageTitle
	short, err := getTemplate(shortTemplatePath)
	if err != nil {
		return err
	}
	var postBlox []string
	for _, post := range posts {
		meta := post.PropSet
		link := fmt.Sprintf("/%s/", post.DirBase)
		ld := ListingData{
			Title:      meta["title"],
			Date:       meta["date"],
			Short:      meta["short"],
			Link:       link,
			Tags:       createTags(meta["tags"]),
			TimeToRead: calculateTimeToRead(post.CntAsHTML),
		}
		execdPostTmplOutput := bytes.Buffer{}
		if err := short.Execute(&execdPostTmplOutput, ld); err != nil {
			return fmt.Errorf("error executing template %s: %v", shortTemplatePath, err)
		}
		postBlox = append(postBlox, execdPostTmplOutput.String())
	}
	htmlBloxFragment := template.HTML(strings.Join(postBlox, "<br />"))
	if g.Config.IsIndex {
		archiveLink, err := getTemplate(archiveLinkTemplatePath)
		if err != nil {
			return err
		}
		execdArchiveLinkTmplOutput := bytes.Buffer{}
		if err := archiveLink.Execute(&execdArchiveLinkTmplOutput, nil); err != nil {
			return fmt.Errorf("error executing template %s: %v", archiveLinkTemplatePath, err)
		}
		htmlBloxFragment = template.HTML(fmt.Sprintf(
			"%s%s", htmlBloxFragment, template.HTML(execdArchiveLinkTmplOutput.String())))
	}
	if err := WriteIndexHTML(g.Config.BlogProps, destDirPath, pageTitle,
			pageTitle, htmlBloxFragment, t); err != nil {
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
