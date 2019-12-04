package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"
	S "strings"
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
	println("Data source type:", dstype)
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
	fmt.Printf("Posts 2b-procest from working-dirs: %v \n", dirs)

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
		return nil, fmt.Errorf(
			"Can't read config file <%s>: %w", "bloggen.yml", e)
	}
	cfgMap, _, e := SU.GetYamlMetadata(string(data))
	if e != nil || cfgMap == nil {
		return nil, fmt.Errorf("Can't parse config: %w", e)
	}
	fmt.Printf("YAML-CFG-MAP: %+v \n", cfgMap)
	ps = make([]SU.PropSet, 3)
	if cfgMap["generator"] == nil ||
	   cfgMap["blog"]     == nil ||
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
		println("Setting default: tmp")
		ps[0]["tmp"] = "tmp"
	}
	if ps[0]["dest"] == "" {
		println("Setting default dest: www")
		ps[0]["dest"] = "www"
	}
	iw := ps[1] // generator.NewIndexWriter(ps)
	if iw["url"] == "" {
		return nil, fmt.Errorf("Please provide a Blog URL, e.g.: https://www.zupzup.org")
	}
	if iw["language"] == "" {
		println("Setting default lg: en-us")
		ps[1]["language"] = "en-us"
	}
	if iw["description"] == "" {
		return nil, fmt.Errorf("Provide a Blog Description, e.g.: A blog about blogging")
	}
	if iw["dateformat"] == "" {
		println("Setting default date format: 02.01.2006")
		ps[1]["dateformat"] = "02.01.2006"
	}
	if iw["title"] == "" {
		return nil, fmt.Errorf("Provide a Blog Title, e.g.: wuzzup")
	}
	if iw["author"] == "" {
		return nil, fmt.Errorf("Provide a Blog author, e.g.: Joe Blow")
	}
	if ps[1]["frontpageposts"] == "0" {
		println("Setting default post count: 10")
		ps[1]["frontpageposts"] = "10"
	}
	return ps, nil
}
