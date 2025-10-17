package pkg

import "fmt"

var (
	Version,
	GitCommit string
)

func UserAgent() string {
	return fmt.Sprintf("alexellis-discord-updater/%s", Version)
}
