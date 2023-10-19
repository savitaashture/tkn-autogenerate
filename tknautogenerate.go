// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-06 09:37:17

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"fmt"
	"strings"

	_ "embed"

	"github.com/alecthomas/kong"
	ag "github.com/chmouel/tknautogenerate/pkg/tknautogenerate"
)

var CLI ag.CliStruct

func main() {
	kong.Parse(&CLI,
		kong.Name("tkn-autogenerate"),
		kong.Description("🧲 Auto generation of pipelinerun on language detection and file patterns"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}))

	output, err := ag.Detect(&CLI)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Println(strings.TrimSpace(output))
}
