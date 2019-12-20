package generator

import (
	"fmt"
	"os"
	S "strings"
	FP "path/filepath"
	GDW "github.com/karrick/godirwalk"
)

// walkFindFs is called after the directory `statics` is copied in full,
// and deletes files that obviously don't belong, but more importantly,
// looks for files named "*.f.*" and applies template processing to them.
func walkFindFs() error {
	dirname := "static"
	dirtgt := "www"
	err := GDW.Walk(dirname, &GDW.Options{
		Callback: func(path string, de *GDW.Dirent) error {
			if !de.IsRegular() { return nil }
			// Delete it ?
			nm := de.Name()
			if S.HasSuffix(nm, "~") {
				nm = FP.Join(dirtgt, nm)
				println("Would delete:", nm)
				return nil
			}
			if S.Contains(nm, ".f.") {
				nm = FP.Join(dirtgt, nm)
				println("Would process & delete:", nm)
				tgt_nm := S.Replace(nm, ".f.", ".", -1)
				 println("Would create:", tgt_nm)
				return nil
			}
			fmt.Printf("%s %s (%s) \n",
				de.ModeType(), path, de.Name())
			return nil
		},
		ErrorCallback: func(path string, err error) GDW.ErrorAction {
			if true { // *optVerbose {
				fmt.Fprintf(os.Stderr, "ERROR on <%s>: %s\n", path, err.Error())
			}

			// For the purposes of this example, a simple SkipNode will suffice,
			// although in reality perhaps additional logic might be called for.
			return GDW.SkipNode
		},
		Unsorted: true, // set true for faster yet non-deterministic enumeration (see godoc)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	return err
}

/*

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

*/
