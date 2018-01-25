package common

import (
	"fmt"
)

var (
	// CommitHash contains the current Git revision. Use make to build to make
	// sure this gets set.
	CommitHash string
	// BuildDate contains the date of the current build.
	BuildDate string
)

var proteusInfo *ProteusInfo

// ProteusVersion contains the version information for proteus
type ProteusVersion struct {
	Major  int
	Minor  int
	Patch  int
	Suffix string
}

func (v ProteusVersion) String() string {
	return proteusVersion(v.Major, v.Minor, v.Patch, v.Suffix)
}

func proteusVersion(major int, minor int, patchVersion int, suffix string) string {
	return fmt.Sprintf("%d.%d.%d%s", major, minor, patchVersion, suffix)
}

// CurrentProteusVersion is the current version of proteus. Remember to change this before making a release
var CurrentProteusVersion = ProteusVersion{
	Major:  0,
	Minor:  2,
	Patch:  0,
	Suffix: "",
}

// ProteusInfo contains information for the current proteus build
type ProteusInfo struct {
	Version    string
	CommitHash string
	BuildDate  string
}

func init() {
	proteusInfo = &ProteusInfo{
		Version:    CurrentProteusVersion.String(),
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
	}
}
