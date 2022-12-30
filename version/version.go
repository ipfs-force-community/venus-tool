package version

var (
	CurrentCommit string

	BuildVersion = "v0.1.0"

	Version = BuildVersion + CurrentCommit
)
