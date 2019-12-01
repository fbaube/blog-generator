package generator

import (
	"fmt"
	SU "github.com/fbaube/stringutils"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	S "strings"
	"strconv"
	"sync"
	"time"
)

// Meta is a data container for per-post Metadata in a "meta.yaml".
type Meta struct {
	Title      string
	Short      string
	Date       string
	Tags       []string
	ParsedDate time.Time
}

// IndexData is a data container for the landing page.
type IndexData struct {
	HTMLTitle       string
	PageTitle       string
	Content         template.HTML
	Year            int
	Name            string
	CanonicalLink   string
	MetaDescription string
	HighlightCSS    template.CSS
}

// Generator interface
type Generator interface {
	Generate() error
}

// SiteGenerator object
type SiteGenerator struct {
	StConfig *SiteConfig
}

// SiteConfig holds the sources and destination folder
type SiteConfig struct {
	Sources     []string
	Dest string
	Configs  []SU.PropSet
}

// New creates a new SiteGenerator.
func New(config *SiteConfig) *SiteGenerator {
	return &SiteGenerator{StConfig: config}
}

// Generate starts the static blog generation.
func (g *SiteGenerator) Generate() error {
	templatePath := filepath.Join("static", "template.html")
	fmt.Println("Generating Site...")
	sources := g.StConfig.Sources
	destination := g.StConfig.Dest
	if err := clearAndCreateDestination(destination); err != nil {
		return err
	}
	if err := clearAndCreateDestination(filepath.Join(destination, "archive")); err != nil {
		return err
	}
	t, err := getTemplate(templatePath)
	if err != nil {
		return err
	}
	var posts []*Post
	for _, path := range sources {
		post, err := newPost(path, g.StConfig.Configs[1]["dateformat"])
		if err != nil {
			return err
		}
		posts = append(posts, post)
	}
	sort.Sort(ByDateDesc(posts))
	if err := runTasks(posts, t, destination, g.StConfig.Configs); err != nil {
		return err
	}
	fmt.Println("Finished generating Site...")
	return nil
}

func runTasks(posts []*Post, t *template.Template, destination string,
		cfgs []SU.PropSet) error {

	var wg sync.WaitGroup
	done := make(chan bool, 1)
	errs := make(chan error, 1)
	pool := make(chan struct{}, 50)
	generators := []Generator{}
	indexWriter := NewIndexWriter(cfgs)
	// ==========================
	//   POSTS
	// ==========================
	for _, post := range posts {
		pPC := new(PostConfig)
		pPC.Post = post
		pPC.Dest = destination
		pPC.Template = t
		pPC.IndexWriter = indexWriter
		println(pPC.String())
		pg := PostGenerator{pPC}
		generators = append(generators, &pg)
	}
	tagPostsMap := createTagPostsMap(posts)
	// ==========================
	//   FRONT PAGE
	// ==========================
	nrposts, _ := strconv.Atoi(cfgs[1]["frontpageposts"])
	pLC := new(ListingConfig)
	pLC.Posts = posts[:getNumOfPagesOnFrontpage(posts, nrposts)]
	pLC.Template = t
	pLC.Dest = destination
	pLC.PageTitle = ""
	pLC.IndexWriter =  indexWriter
	pLC.IsIndex = true
	println(pLC.String())
	fg := ListingGenerator{pLC}
	// ==========================
	//   ARCHIVE
	// ==========================
	pAC := new(ListingConfig)
	pAC.Posts = posts
	pAC.Template = t
	pAC.Dest = filepath.Join(destination, "archive")
	pAC.PageTitle = "Archive"
	pAC.IndexWriter =  indexWriter
	pAC.IsIndex = false
	println(pAC.String())
	ag := ListingGenerator{pAC}
	// ==========================
	//   TAGS
	// ==========================
	pTC := new(TagsConfig)
	pTC.TagPostsMap = tagPostsMap
	pTC.Template = t
	pTC.Dest = destination
	pTC.IndexWriter =  indexWriter
	println("TagsCfg:", pTC.BaseConfig.String(),
		fmt.Sprintf("; \n\t TagPostsMap: %+v", pTC.TagPostsMap))
	tg := TagsGenerator{pTC}
	staticURLs := []string{}
	var file, tmpl string
	var files, tmpls []string
	file = cfgs[2]["files"]
	tmpl = cfgs[2]["templates"]
	println("FILEs:", file)
	println("TMPLs:", tmpl)
	files = S.Split(file, " ")
	tmpls = S.Split(tmpl, " ")
	fmt.Printf("FILES: %v \n", files)
	fmt.Printf("TMPLS: %v \n", tmpls)

	for _, staticURL := range tmpls {
		staticURLs = append(staticURLs, staticURL) // .Dest)
	}
	// sitemap
	sg := SitemapGenerator{&SitemapConfig{
		Posts:       posts,
		TagPostsMap: tagPostsMap,
		Destination: destination,
		BlogURL:     cfgs[1]["url"],
		Statics:     staticURLs,
	}}
	// rss
	rg := RSSGenerator{&RSSConfig{
		Posts:           posts,
		Destination:     destination,
		DateFormat:      cfgs[1]["dateformat"],
		Language:        cfgs[1]["language"],
		BlogURL:         cfgs[1]["url"],
		BlogDescription: cfgs[1]["description"],
		BlogTitle:       cfgs[1]["title"],
	}}
	// ==========================
	//   STATICS
	// ==========================
	fileToDestination := map[string]string{}
	for _, static := range files {
		fileToDestination["static/" + static] = filepath.Join(destination, static) // .Dest)
	}
	// ==========================
	//   TEMPLATES
	// ==========================
	templateToFile := map[string]string{}
	for _, static := range tmpls {
		templateToFile["static/" + static] = filepath.Join(destination, static, "index.html")
	}
	pSC := new(StaticsConfig)
	pSC.FilesToDests = fileToDestination
	pSC.TmplsToFiles = templateToFile
	pSC.Template = t
	pSC.IndexWriter = indexWriter
	fmt.Printf("StcsCfg: %s; \n\t %+v %+v \n",
		pSC.BaseConfig.String(), pSC.FilesToDests, pSC.TmplsToFiles)
	statg := StaticsGenerator{pSC}
	generators = append(generators, &fg, &ag, &tg, &sg, &rg, &statg)

	for _, generator := range generators {
		wg.Add(1)
		go func(g Generator) {
			defer wg.Done()
			pool <- struct{}{}
			defer func() { <-pool }()
			if err := g.Generate(); err != nil {
				errs <- err
			}
		}(generator)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case err := <-errs:
		if err != nil {
			return err
		}
	}
	return nil
}

func clearAndCreateDestination(path string) error {
	if err := os.RemoveAll(path); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("error removing folder at destination %s: %v ", path, err)
		}
	}
	return os.Mkdir(path, os.ModePerm)
}

func getHTMLTitle(pageTitle, blogTitle string) string {
	if pageTitle == "" {
		return blogTitle
	}
	return fmt.Sprintf("%s - %s", pageTitle, blogTitle)
}

func createTagPostsMap(posts []*Post) map[string][]*Post {
	result := make(map[string][]*Post)
	for _, post := range posts {
		for _, tag := range post.Meta.Tags {
			key := S.ToLower(tag)
			if result[key] == nil {
				result[key] = []*Post{post}
			} else {
				result[key] = append(result[key], post)
			}
		}
	}
	return result
}

func getTemplate(path string) (*template.Template, error) {
	t, err := template.ParseFiles(path)
	if err != nil {
		return nil, fmt.Errorf("error reading template %s: %v", path, err)
	}
	return t, nil
}

func getNumOfPagesOnFrontpage(posts []*Post, numPosts int) int {
	if len(posts) < numPosts {
		return len(posts)
	}
	return numPosts
}

func buildCanonicalLink(path, baseURL string) string {
	parts := S.Split(path, "/")
	if len(parts) > 2 {
		return fmt.Sprintf("%s/%s/index.html", baseURL, S.Join(parts[2:], "/"))
	}
	return "/"
}
