package generator

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	FP "path/filepath"
	"strings"
  FU "github.com/fbaube/fileutils"
)

// StaticsGenerator object
type StaticsGenerator struct {
	Config *StaticsConfig
}

// StaticsConfig holds the data for the static sites.
type StaticsConfig struct {
	TmplsToFiles map[string]string
	BaseConfig
}

// Generate creates the static pages.
func (g *StaticsGenerator) Generate() error {
	fmt.Println("\tCopying Statics...")
	psTmplsToFiles := g.Config.TmplsToFiles
	fmt.Printf("StcsGenr: TmplsToFiles: %+v \n", psTmplsToFiles)
	t := g.Config.Template
	// Both arguments should be directories !!
	// func CopyDirRecursivelyFromTo(srcFrom string, dstTo string) error {
  if err := FU.CopyDirRecursivelyFromTo("static", g.Config.Dest); err != nil {
		return err
	}

	for k, v := range psTmplsToFiles {
		if err := createFolderIfNotExist(FP.Dir(v)); err != nil {
			return err
		}
		content, err := ioutil.ReadFile(k)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", k, err)
		}
		println("Calling WriteIndexHTML:", k, v)
		if err := WriteIndexHTML(g.Config.BlogProps, FP.Dir(v), getTitle(k), getTitle(k), template.HTML(content), t); err != nil {
			return err
		}
	}
	fmt.Println("\tFinished copying statics...")
	return nil
}

func createFolderIfNotExist(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err = os.Mkdir(path, os.ModePerm); err != nil {
				return fmt.Errorf("error creating directory %s: %v", path, err)
			}
		} else {
			return fmt.Errorf("error accessing directory %s: %v", path, err)
		}
	}
	return nil
}

func getTitle(path string) string {
	ext := FP.Ext(path)
	name := FP.Base(path)
	fileName := name[:len(name)-len(ext)]
	return fmt.Sprintf("%s%s", strings.ToUpper(string(fileName[0])), fileName[1:])
}
