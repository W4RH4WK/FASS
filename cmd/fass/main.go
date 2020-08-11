package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/W4RH4WK/FASS/pkg/fass"
)

const mappingFilename = "mapping.json"

func printUsage() {
	const usage = `usage: %s <command>

commands:

  serve                                            Start FASS service.
  token <mail-file>                                Generate a token for each mail address and produces a '%s' file.
  course <identifier> <mapping-file>               Generate a course, adding the tokens from the given mapping file.
  distribute <course-identifier> <mapping-file>    Distributes the generated tokens via mail.

config:

  Config file is located at ~/.config/fass/config.json.
`

	fmt.Fprintf(os.Stderr, usage, os.Args[0], mappingFilename)
	flag.PrintDefaults()
}

func generateTokenMapping(mailFilepath string) {
	mailFile, err := os.Open(mailFilepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer mailFile.Close()

	mapping := fass.NewTokenMapping(mailFile)
	mapping.Store(mappingFilename)
}

func generateCourse(identifier string, mappingFilepath string) {
	_, err := fass.LoadCourse(identifier)
	if err == nil {
		fmt.Fprintln(os.Stderr, "course already exists")
		return
	}

	mapping, err := fass.LoadTokenMapping(mappingFilepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	course := fass.Course{
		Identifier: identifier,
		Name: "course name",
		URL: "http://example.org",
		Path: identifier,
	}

	for token := range mapping {
		course.Users = append(course.Users, token)
	}

	err = course.Store()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
}

func distributeTokens(config fass.Config, courseIdentifier string, mappingFilepath string) {
	course, err := fass.LoadCourse(courseIdentifier)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	mapping, err := fass.LoadTokenMapping(mappingFilepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	for _, user := range course.Users {
		if addr, found := mapping[user]; found {
			fmt.Println("sending:", addr)
			err := fass.DistributeToken(user, addr, course, config)
			if err != nil {
				fmt.Fprintln(os.Stderr, addr, err.Error())
			}
		}
	}
}

func loadConfig() fass.Config {
	config, err := fass.LoadConfig()
	if os.IsNotExist(err) {
		config = fass.DefaultConfig()
		err = config.Store()
	}

	if err != nil {
		panic(err.Error())
	}

	return config
}

func warnAboutMapping() {
	_, err := os.Stat(mappingFilename)
	if os.IsNotExist(err) {
		return
	}

	fmt.Fprintf(os.Stderr, "Warning: %s present\n", mappingFilename)
}

func main() {
	config := loadConfig()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "serve":
		warnAboutMapping()
		fass.Serve(config.ListenAddress)
	case "token":
		generateTokenMapping(os.Args[2])
	case "course":
		generateCourse(os.Args[2], os.Args[3])
	case "distribute":
		distributeTokens(config, os.Args[2], os.Args[3])
	case "help":
		printUsage()
		os.Exit(0)
	default:
		printUsage()
		os.Exit(2)
	}
}
