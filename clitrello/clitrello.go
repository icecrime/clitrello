package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func appName() (name string) {
	if name = os.Args[0]; strings.HasPrefix(name, "./") {
		name = name[2:]
	}
	return
}

func configurationShouldBePersisted(config *Config) bool {
	fmt.Print("Do you want to save this token? It will be written in clear text to '", config.File, "' [Y/N] ")
	answer := readLine()
	return answer == "y" || answer == "Y"
}

func readLine() string {
	input := ""
	if _, err := fmt.Scanln(&input); err != nil {
		log.Fatal(err)
	}
	return input
}

func requestApiTokenWorkflow(config *Config) string {
	// Printing the URL to navigate to sucks, but Goland doesn't seem to have
	// an equivalent of Python's webbrowser module. To be continued...
	fmt.Printf(`Follow these steps to obtain an authentication token:
  1. Go to: %s
  2. Authorize access in order to obtain a token.
  3. Enter provided token: `, AuthorizationURL(config, appName()))
	return readLine()
}

/**
 ******************************************************************************
 */

func main() {
	// There's not much we can do without a user token but hey, why not.
	config := GetConfiguration()
	if config.Token == "" {
		fmt.Println("warning: no user token provided (only public boards are accessible)")
	}

	switch cmd := flag.Arg(0); cmd {
	case "init":
		config.Token = requestApiTokenWorkflow(config)
		if configurationShouldBePersisted(config) {
			config.Save(config.File)
			fmt.Println("Configuration saved.")
		}
	case "index":
		Visual(config)
	default:
		fmt.Print("invalid command '", cmd, "'\n")
	}
}
