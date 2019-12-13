package datasource

import (
	"fmt"
	S "strings"
	"path"
	FU "github.com/fbaube/fileutils"
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
	fmt.Printf("Fetching data from %s \n              into %s...\n", from, to)
	var e error

	if e = FU.MakeDirectoryExist(to); e != nil {
		return nil, e
	}
	if e = FU.ClearDirectory(to); e != nil {
		return nil, e
	}
	// if err := cloneFileDir(to, from); err != nil {
	if e = FU.CopyDirRecursivelyFromTo(from, to); e != nil {
		return nil, e
	}
	var dirs []string
	dirs, e = getContentFolders(to)
	if e != nil {
		return nil, e
	}
	// fmt.Print("Fetching complete.\n")
	return dirs, nil
}
