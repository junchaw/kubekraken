package main

import (
	"github.com/junchaw/kubekraken/cmd"
)

func main() {
	rootCmd := cmd.NewKrakenCmd()

	rootCmd.Execute()
}
