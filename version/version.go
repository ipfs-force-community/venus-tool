package version

var (
	CurrentCommit string

	BuildVersion = "v0.1.3"

	Version = BuildVersion + CurrentCommit
)
