package cli

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	S "strings"
	"path"
	FU "github.com/fbaube/fileutils"
	"github.com/fbaube/bloggenator/config"
	"github.com/fbaube/bloggenator/datasource"
	"github.com/fbaube/bloggenator/generator"
)

// Run runs the application. Amaze!
func Run() {
	cfg, err := readConfig()
	if err != nil {
		log.Fatal("Can't read configuration file: ", err)
	}
	var dstype string
	var repo string
	repo = cfg.Generator.Repo
	hasProtocol := S.Contains(repo, "://")
	idxProtocol := S.Index(repo, "://")
	if hasProtocol && S.HasPrefix(repo, "http") {
		fmt.Printf("Repo protocol is %s... \n", repo[:idxProtocol+3])
		dstype = "git"
	} else if S.HasPrefix(repo, "file://") || path.IsAbs(repo) {
		fmt.Printf("Repo is a directory... \n")
		dstype = "filesystem"
	} else {
		log.Fatal(fmt.Errorf("unknown protocol: %s", repo))
	}
	// Check that arguments are OK
	var chp_tmpTo, chp_repo *FU.CheckedPath
  var tmpTo string
	tmpTo = cfg.Generator.Tmp
	chp_tmpTo = FU.NewCheckedPath(tmpTo)
	if chp_tmpTo.Exists && !chp_tmpTo.IsDir {
		log.Fatal(fmt.Errorf("\"Tmp\" is not a directory: <%s>", tmpTo))
	}
	if dstype == "filesystem" {
		chp_repo = FU.NewCheckedPath(repo)
		if !(chp_repo.Exists && chp_repo.IsDir) {
			log.Fatal(fmt.Errorf("HTTP Repo is not a directory: <%s>", repo))
		}
	}
	ds := datasource.New(dstype)
	var dirs []string
	// Fetch(from, to string)
	switch dstype {
	case "git":
		dirs, err = ds.Fetch(cfg.Generator.Repo, cfg.Generator.Tmp)
	case "filesystem":
		dirs, err = ds.Fetch(chp_repo.AbsFilePath.S(), chp_tmpTo.AbsFilePath.S())
	}
	if err != nil {
		log.Fatal(err)
	}
	g := generator.New(&generator.SiteConfig{
		Sources:     dirs,
		Destination: cfg.Generator.Dest,
		Config:      cfg,
	})

	err = g.Generate()

	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() (*config.Config, error) {
	data, err := ioutil.ReadFile("bloggen.yml")
	if err != nil {
		return nil, fmt.Errorf("Can't read config file: %w", err)
	}
	cfg := config.Config{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("Can't parse config: %w", err)
	}
	if cfg.Generator.Repo == "" {
		return nil, fmt.Errorf("Provide a repository URL: a directory filepath or http[s]")
	}
	if cfg.Generator.Tmp == "" {
		cfg.Generator.Tmp = "tmp"
	}
	if cfg.Generator.Dest == "" {
		cfg.Generator.Dest = "www"
	}
	if cfg.Blog.URL == "" {
		return nil, fmt.Errorf("Please provide a Blog URL, e.g.: https://www.zupzup.org")
	}
	if cfg.Blog.Language == "" {
		cfg.Blog.Language = "en-us"
	}
	if cfg.Blog.Description == "" {
		return nil, fmt.Errorf("Provide a Blog Description, e.g.: A blog about blogging")
	}
	if cfg.Blog.Dateformat == "" {
		cfg.Blog.Dateformat = "02.01.2006"
	}
	if cfg.Blog.Title == "" {
		return nil, fmt.Errorf("Provide a Blog Title, e.g.: wuzzup")
	}
	if cfg.Blog.Author == "" {
		return nil, fmt.Errorf("Provide a Blog author, e.g.: Joe Blow")
	}
	if cfg.Blog.Frontpageposts == 0 {
		cfg.Blog.Frontpageposts = 10
	}
	return &cfg, nil
}
