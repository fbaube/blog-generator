package datasource

import (
	"fmt"
)
// DataSource is the data-source fetching interface
type DataSource interface {
	Fetch(from, to string) ([]string, error)
}

// New creates a new DataSource
func New(dstype string) DataSource {
	fmt.Printf("Repo datasource type is: %s \n", dstype)
	switch dstype {
	case "git":
		return &GitDataSource{}
	case "filesystem":
		return &FileSystemDataSource{}
	default:
		return nil
	}
}
