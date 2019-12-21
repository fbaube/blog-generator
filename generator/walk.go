package generator

import (
	"fmt"
	"os"
	"io/ioutil"
	S "strings"
	// FP "path/filepath"
	"html/template"
	SU "github.com/fbaube/stringutils"
	GDW "github.com/karrick/godirwalk"
	"github.com/morningconsult/serrors"
)

// walkFindFs is called after the directory `statics` is copied in full,
// and deletes files that obviously don't belong, but more importantly,
// looks for files named "*.f.*" and applies template processing to them.
func walkFindFs(dest string, t *template.Template, blogProps SU.PropSet) error {
	println("walkFindFs:", "dest:", dest)
	err := GDW.Walk(dest, &GDW.Options{
		Callback: func(relpath string, de *GDW.Dirent) error {
			if !de.IsRegular() { return nil }

			if S.HasSuffix(relpath, "~") {
				println("Deleting:", relpath)
				e := os.Remove(relpath)
				if e != nil {
					println ("rm failed:", relpath, e.Error())
					}
				return nil
			}
			// Delete it ?
			fnm := de.Name()
			// fmt.Printf("path <%s> name<%s> \n", relpath, fnm)
			// dirpath := S.TrimSuffix(relpath, fnm)

			if S.Contains(fnm, ".f.") {
				println("Would process & delete:", relpath)
				newpath := S.Replace(relpath, ".f.", ".", -1)
				println("Would create:", newpath)
				e := duit(relpath, newpath, t, blogProps)
				if e != nil {
					println ("duit failed:", relpath, newpath, e.Error())
					}
				e = os.Remove(relpath)
				if e != nil {
					println ("rm failed:", relpath, e.Error())
					}
				return nil
			}
			fmt.Printf("%s %s (%s) \n",
				de.ModeType(), relpath, de.Name())
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

func duit(oldpath, newpath string, t *template.Template, bp SU.PropSet) error {

	content, err := ioutil.ReadFile(oldpath)
	if err != nil {
		return serrors.Errorf("error reading file %s: %w", oldpath, err)
	}
	println("Calling WrapHtmlInMasterPageTemplate:", oldpath, newpath)
	// WriteIndexHTML(blogProps SU.PropSet, destDirPath, pageTitle,
	// aMetaDesc string, htmlContentFrag template.HTML, t *template.Template)
	targs := *new(IndexHtmlMasterPageTemplateVariableArguments)
	targs.PageTitle = getTitle(oldpath)
	targs.HtmlTitle = getTitle(oldpath)
	targs.HtmlContentFrag = template.HTML(content)
	if err := WrapHtmlInMasterPageTemplate(targs, bp, newpath, t); err != nil {
		return err
	}
	return nil
}
