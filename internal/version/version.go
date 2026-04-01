package version

import "fmt"

const Name = "blockgo"

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func String() string {
	return fmt.Sprintf("%s %s (commit=%s date=%s)", Name, Version, Commit, Date)
}
