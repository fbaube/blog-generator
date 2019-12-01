package datasource

import (
	"fmt"
	"os"
	"io"
	"io/ioutil"
	// "os/exec"
	S "strings"
	"path"
	// FU "github.com/fbaube/fileutils"
)

// FileSystemDataSource is the file system data source object.
type FileSystemDataSource struct{}

// Fetch creates the output folder, clears it and copies the repository there.
func (ds *FileSystemDataSource) Fetch(from, to string) ([]string, error) {
	if S.Contains(from, "://") && !S.HasPrefix(from, "file://") {
		panic("bad protocol: " + from)
	}
	if !path.IsAbs(from) {
		panic("not absolute filepath: " + from)
	}
	fmt.Printf("Fetching data from %s into %s...\n", from, to)
	var e error

	if e = createFolderIfNotExist(to); e != nil {
		return nil, e
	}
	if e = clearFolder(to); e != nil {
		return nil, e
	}
	// if err := cloneFileDir(to, from); err != nil {
	if e = CopyDirRecursivelyFromTo(from, to); e != nil {
		return nil, e
	}
	var dirs []string
	dirs, e = getContentFolders(to)
	if e != nil {
		return nil, e
	}
	fmt.Print("Fetching complete.\n")
	return dirs, nil
}

func cloneFileDir(pathTo string, repoFrom string) error {
	b := CopyDirRecursivelyFromTo(repoFrom, pathTo)
	return b
}

// CopyDirRecursivelyFromTo copies a whole directory recursively.
// Both argument should be directories !!
func CopyDirRecursivelyFromTo(srcFrom string, dstTo string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(srcFrom); err != nil {
		return err
	}
	if err = os.MkdirAll(dstTo, srcinfo.Mode()); err != nil {
		return err
	}
	if fds, err = ioutil.ReadDir(srcFrom); err != nil {
		return err
	}
	println("src:", srcFrom)
	println("dst:", dstTo)

	for _, fd := range fds {
		srcfp := path.Join(srcFrom, fd.Name())
		dstfp := path.Join(dstTo, fd.Name())

		if fd.IsDir() {
			if err = CopyDirRecursivelyFromTo(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = CopyFileFromTo(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
}

// CopyFileFromTo copies a single file from src to dst.
func CopyFileFromTo(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}
