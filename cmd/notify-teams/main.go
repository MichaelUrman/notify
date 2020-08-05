/*

Command Notify is a small webhook client that forms and posts a Microsof Teams
webhook request to notify about GitHub workflow events.

*/
package main

import (
	"os"

	"github.com/MichaelUrman/notify/internal/github"
	"github.com/MichaelUrman/notify/internal/notifier"
	"github.com/MichaelUrman/notify/internal/teams"
)

func main() {
	err := notifier.Main(
		github.LoadEvent,
		teams.BuildSubmitter,
	)
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
