package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/mitchellh/cli"
)

// helpFunc is a cli.HelpFunc that can is used to output the help for Terraform.
func helpFunc(commands map[string]cli.CommandFactory) string {
	return "Help is disabled since this is a modded version."
}

// listCommands just lists the commands in the map with the
// given maximum key length.
func listCommands(allCommands map[string]cli.CommandFactory, order []string, maxKeyLen int) string {
	var buf bytes.Buffer

	for _, key := range order {
		commandFunc, ok := allCommands[key]
		if !ok {
			// This suggests an inconsistency in the command table definitions
			// in commands.go .
			panic("command not found: " + key)
		}

		command, err := commandFunc()
		if err != nil {
			// This would be really weird since there's no good reason for
			// any of our command factories to fail.
			log.Printf("[ERR] cli: Command '%s' failed to load: %s",
				key, err)
			continue
		}

		key = fmt.Sprintf("%s%s", key, strings.Repeat(" ", maxKeyLen-len(key)))
		buf.WriteString(fmt.Sprintf("  %s  %s\n", key, command.Synopsis()))
	}

	return buf.String()
}
