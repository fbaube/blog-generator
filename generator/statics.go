package generator

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	FP "path/filepath"
	"strings"
  FU "github.com/fbaube/fileutils"
	"github.com/morningconsult/serrors"
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
	// TODO Treewalk to find ".f." files
	_ = walkFindFs()

	fmt.Printf("Nr of TmplsToFiles: %d \n", len(psTmplsToFiles))
	for k, v := range psTmplsToFiles {
		if err := createFolderIfNotExist(FP.Dir(v)); err != nil {
			return err
		}
		content, err := ioutil.ReadFile(k)
		if err != nil {
			return serrors.Errorf("error reading file %s: %w", k, err)
		}
		println("Calling WriteIndexHTML:", k, v)
		// WriteIndexHTML(blogProps SU.PropSet, destDirPath, pageTitle,
		// aMetaDesc string, htmlContentFrag template.HTML, t *template.Template)
		targs := *new(IndexHtmlMasterPageTemplateVariableArguments)
		targs.PageTitle = getTitle(k)
		targs.HtmlTitle = getTitle(k)
		targs.HtmlContentFrag = template.HTML(content)
		if err := WriteIndexHTML(targs, g.Config.BlogProps, FP.Dir(v), t); err != nil {
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
				return serrors.Errorf("error creating directory %s: %w", path, err)
			}
		} else {
			return serrors.Errorf("error accessing directory %s: %w", path, err)
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
