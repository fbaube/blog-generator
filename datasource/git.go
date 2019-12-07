package datasource

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	S "strings"
	"github.com/morningconsult/serrors"
	FU "github.com/fbaube/fileutils"
)

// GitDataSource is the git data source object
type GitDataSource struct{}

// Fetch creates the output folder, clears it and clones the repository there
func (ds *GitDataSource) Fetch(from, to string) ([]string, error) {
	if !S.Contains(from, "://") {
		panic("missing protocol")
	}
	if !S.HasPrefix(from, "http") {
		panic("bad protocol")
	}
	fmt.Printf("Fetching data into %s from %s ...\n", to, from)

	if err := FU.MakeDirectoryExist(to); err != nil {
		return nil, err
	}
	if err := FU.ClearDirectory(to); err != nil {
		return nil, err
	}
	if err := cloneRepo(to, from); err != nil {
		return nil, err
	}
	dirs, err := getContentFolders(to)
	if err != nil {
		return nil, err
	}
	fmt.Print("Fetching complete.\n")
	return dirs, nil
}

func cloneRepo(path, repositoryURL string) error {
	cmdName := "git"
	initArgs := []string{"init", "."}
	cmd := exec.Command(cmdName, initArgs...)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return serrors.Errorf("error initializing git repository at %s: %w", path, err)
	}
	remoteArgs := []string{"remote", "add", "origin", repositoryURL}
	cmd = exec.Command(cmdName, remoteArgs...)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return serrors.Errorf("error setting remote %s: %w", repositoryURL, err)
	}
	pullArgs := []string{"pull", "origin", "master"}
	cmd = exec.Command(cmdName, pullArgs...)
	cmd.Dir = path
	if err := cmd.Run(); err != nil {
		return serrors.Errorf("error pulling master at %s: %w", path, err)
	}
	return nil
}

func getContentFolders(path string) ([]string, error) {
	var result []string
	dir, err := os.Open(path)
	if err != nil {
		return nil, serrors.Errorf("error accessing directory %s: %w", path, err)
	}
	defer dir.Close()
	files, err := dir.Readdir(-1)
	if err != nil {
		return nil, serrors.Errorf("error reading contents of directory %s: %w", path, err)
	}
	for _, file := range files {
		if file.IsDir() && file.Name()[0] != '.' {
			result = append(result, filepath.Join(path, file.Name()))
		}
	}
	return result, nil
}
