package generator

import (
	"fmt"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
	"html/template"
	"path/filepath"
	"sort"
	S "strings"
	"strconv"
	"sync"
	"github.com/morningconsult/serrors"
)

// DO_SEQUENTIALLY simplifies debugging by forcing all Generators
// to run sequentially - not concurrently - in func runTasks()
var DO_SEQUENTIALLY = true

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
	Sources    []string
	Dest         string
	Configs []SU.PropSet
}

// New creates a new SiteGenerator.
func New(config *SiteConfig) *SiteGenerator {
	return &SiteGenerator{StConfig: config}
}

// Generate starts the static blog generation.
func (g *SiteGenerator) Generate() error {
	masterPageTmplPath := filepath.Join("template", "masterpagetemplate.html")
	sources := g.StConfig.Sources
	destination := g.StConfig.Dest
	fmt.Printf("SiteGenr: dest<%s> \n", destination)
	// fmt.Printf("          sources: %+v \n", sources)

	// Clear the WWW output directory and its "archive" subdirectory
	if err := FU.ClearAndCreateDirectory(destination); err != nil {
		return err
	}
	if err := FU.ClearAndCreateDirectory(filepath.Join(destination, "archive")); err != nil {
		return err
	}

	// Get the master page template
	masterPageTemplate, err := getTemplate(masterPageTmplPath)
	if err != nil {
		return err
	}

	// Get all the posts
	var posts []*Post
	for _, path := range sources {
		println("##>> Adding post, from wrkg-dir:", path)
		post, err := newPost(path, g.StConfig.Configs[1]["dateformat"])
		if err != nil {
			return err
		}
		posts = append(posts, post)
	}
	sort.Sort(ByDateDesc(posts))
	if err := runTasks(posts, masterPageTemplate, destination, g.StConfig.Configs); err != nil {
		return err
	}
	fmt.Println("Finished generating Site...")
	return nil
}

func runTasks(posts []*Post, masterPageTemplate *template.Template,
		destination string, cfgs []SU.PropSet) error {

	generators := []Generator{}
	blogProps := cfgs[1]
	// ==========================
	//   POSTS
	// ==========================
	for _, post := range posts {
		pPC := new(PostConfig)
		pPC.Post = post
		pPC.Dest = destination
		pPC.Template = masterPageTemplate
		pPC.BlogProps = blogProps
		// println(pPC.String())
		pg := PostGenerator{pPC}
		generators = append(generators, &pg)
		// fmt.Printf("##>> Ready post: %+v |||| \n", *pPC)
	}
	tagPostsMap := createTagPostsMap(posts)
	// ==========================
	//   FRONT PAGE
	// ==========================
	nrposts, _ := strconv.Atoi(cfgs[1]["frontpageposts"])
	pLC := new(ListingConfig)
	pLC.Posts = posts[:getNumOfPagesOnFrontpage(posts, nrposts)]
	pLC.Template = masterPageTemplate
	pLC.Dest = destination
	pLC.PageTitle = ""
	pLC.BlogProps = blogProps
	pLC.IsIndex = true
	// println(pLC.String())
	fg := ListingGenerator{pLC}
	// ==========================
	//   ARCHIVE
	// ==========================
	pAC := new(ListingConfig)
	pAC.Posts = posts
	pAC.Template = masterPageTemplate
	pAC.Dest = filepath.Join(destination, "archive")
	pAC.PageTitle = "Archive"
	pAC.BlogProps = blogProps
	pAC.IsIndex = false
	// println(pAC.String())
	ag := ListingGenerator{pAC}
	// ==========================
	//   TAGS
	// ==========================
	pTC := new(TagsConfig)
	pTC.TagPostsMap = tagPostsMap
	pTC.Template = masterPageTemplate
	pTC.Dest = destination
	pTC.BlogProps = blogProps
	println("TagsCfg:", pTC.BaseConfig.String(),
		fmt.Sprintf("; \n\t TagPostsMap: %+v", pTC.TagPostsMap))
	tg := TagsGenerator{pTC}
	staticURLs := []string{}
	var /* file, */ tmpl string
	var /* files, */ tmpls []string
	// file = cfgs[2]["files"]
	tmpl = cfgs[2]["templates"]
	// files = S.Split(file, " ")
	tmpls = S.Split(tmpl, " ")
	// fmt.Printf("FILES: %v \n", files)
	fmt.Printf("TMPLS: %v \n", tmpls)
	for _, staticURL := range tmpls {
		staticURLs = append(staticURLs, staticURL) // .Dest)
	}
	// ==========================
	//   SITEMAP
	// ==========================
	sg := SitemapGenerator{&SitemapConfig{
		Posts:       posts,
		TagPostsMap: tagPostsMap,
		Destination: destination,
		BlogURL:     cfgs[1]["url"],
		Statics:     staticURLs,
	}}
	// ==========================
	//   RSS
	// ==========================
	pRC := new(RSSConfig)
	pRC.BlogProps = cfgs[1]
	pRC.Posts      = posts
	pRC.Dest       = destination
	rg := RSSGenerator{pRC}
	// ==========================
	//   STATICS
	// ==========================
	pSC := new(StaticsConfig)
	// ==========================
	//   TEMPLATES
	// ==========================
	psTmplsToFiles := SU.PropSet{} // map[string]string{}
	for _, staticFrom := range tmpls {
		staticTo := S.TrimSuffix(staticFrom, ".html")
		staticTo = S.TrimSuffix(staticTo, ".htm")
		psTmplsToFiles["static/" + staticFrom] = filepath.Join(destination, staticTo, "index.html")
	}
	// ==========================
	//   TEMPLATES
	// ==========================
	pSC.TmplsToFiles = psTmplsToFiles
	pSC.Dest = cfgs[0]["dest"]
	pSC.Template = masterPageTemplate
	pSC.BlogProps = blogProps
	fmt.Printf("StcsCfg: %s; \n\t %+v \n",
		pSC.BaseConfig.String(), pSC.TmplsToFiles) 
	statg := StaticsGenerator{pSC}
	generators = append(generators, &fg, &ag, &tg, &sg, &rg, &statg)

	if DO_SEQUENTIALLY {
		for _, g := range generators {
			if err := g.Generate(); err != nil {
				println("ERROR:", err.Error())
				return err
			}
		}
		return nil
	}

	var wg sync.WaitGroup
	done := make(chan bool, 1)
	errs := make(chan error, 1)
	pool := make(chan struct{}, 50)

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

func createTagPostsMap(posts []*Post) map[string][]*Post {
	result := make(map[string][]*Post)
	for _, post := range posts {
		tags := S.Split(post.PropSet["tags"], " ")
		for _, tag := range tags {
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
		return nil, serrors.Errorf("error reading template %s: %w", path, err)
	}
	return t, nil
}

func getNumOfPagesOnFrontpage(posts []*Post, numPosts int) int {
	if len(posts) < numPosts {
		return len(posts)
	}
	return numPosts
}

// buildCanonicalLink builds a URL by taking the baseURL and appending
// the path, where the first two segments of the path are dropped.
func buildCanonicalLink(baseURL, path string) string {
	parts := S.Split(path, "/")
	if len(parts) > 2 {
		retstr := fmt.Sprintf("%s/%s/index.html", baseURL, S.Join(parts[2:], "/"))
		fmt.Printf("##>> bldCanon: baseURL<%s> path <%s> => Canon<%s> \n",
			baseURL, path, retstr)
		return retstr
	}
	return "/"
}
