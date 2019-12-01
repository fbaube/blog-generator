package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	S "strings"
	"path"
	FU "github.com/fbaube/fileutils"
	SU "github.com/fbaube/stringutils"
	"github.com/fbaube/bloggenator/datasource"
	"github.com/fbaube/bloggenator/generator"
)

// Run runs the application. Amaze!
func Run() {
	cfgs, err := readConfig()
	if err != nil {
		log.Fatal("Can't read configuration file: ", err)
	}
	var dstype string
	var repo string
	repo = cfgs[0]["repo"]
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
	println("DSTYPE:", dstype)
	// Check that arguments are OK
	var chp_tmpTo, chp_repo *FU.CheckedPath
  var tmpTo string
	tmpTo = cfgs[0]["tmp"]
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
		dirs, err = ds.Fetch(cfgs[0]["repo"], cfgs[0]["tmp"])
	case "filesystem":
		dirs, err = ds.Fetch(chp_repo.AbsFilePath.S(), chp_tmpTo.AbsFilePath.S())
	}
	if err != nil {
		log.Fatal(err)
	}
	g := generator.New(&generator.SiteConfig{
		Sources: dirs,
		Dest:    cfgs[0]["dest"],
		Configs: cfgs,
	})

	err = g.Generate()

	if err != nil {
		log.Fatal(err)
	}
}

func readConfig() (ps []SU.PropSet, e error) {
	data, e := ioutil.ReadFile("bloggen.yml")
	if e != nil {
		return nil, fmt.Errorf("Can't read config file: %w", e)
	}
	cfgMap, _, e := SU.GetYamlMetadata(string(data))
	if e != nil || cfgMap == nil {
		return nil, fmt.Errorf("Can't parse config: %w", e)
	}
	fmt.Printf("CFG-MAP: %+v \n", cfgMap)

	ps = make([]SU.PropSet, 3)
	if cfgMap["generator"] == nil ||
	   cfgMap["blog"] == nil ||
		 cfgMap["statics"] == nil {
			panic("nil's in cli.go")
		}
	ps[0] = SU.YamlMapAsPropSet(cfgMap["generator"].(map[interface{}]interface{}))
	ps[1] = SU.YamlMapAsPropSet(cfgMap["blog"].(map[interface{}]interface{}))
	ps[2] = SU.YamlMapAsPropSet(cfgMap["statics"].(map[interface{}]interface{}))
	if ps[0]["repo"] == "" {
		return nil, fmt.Errorf("Provide a repo URL: filepath:// or http[s]://")
	}
	if ps[0]["tmp"] == "" {
		ps[0]["tmp"] = "tmp"
	}
	if ps[0]["dest"] == "" {
		ps[0]["dest"] = "www"
	}
	if ps[1]["url"] == "" {
		return nil, fmt.Errorf("Please provide a Blog URL, e.g.: https://www.zupzup.org")
	}
	if ps[1]["language"] == "" {
		ps[1]["language"] = "en-us"
	}
	if ps[1]["description"] == "" {
		return nil, fmt.Errorf("Provide a Blog Description, e.g.: A blog about blogging")
	}
	if ps[1]["dateformat"] == "" {
		ps[1]["dateformat"] = "02.01.2006"
	}
	if ps[1]["title"] == "" {
		return nil, fmt.Errorf("Provide a Blog Title, e.g.: wuzzup")
	}
	if ps[1]["author"] == "" {
		return nil, fmt.Errorf("Provide a Blog author, e.g.: Joe Blow")
	}
	if ps[1]["frontpageposts"] == "0" {
		ps[1]["frontpageposts"] = "10"
	}
	/*
	iw = new(generator.IndexWriter)
	iw.BlogTitle  = ps[1]["title"]
	iw.BlogDesc   = ps[1]["description"]
	iw.BlogAuthor = ps[1]["author"]
	iw.BlogURL    = ps[1]["url"]
	*/
	return ps, nil
}
