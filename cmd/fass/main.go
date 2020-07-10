package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/W4RH4WK/FASS/pkg/fass"
)

func printUsage() {
	const usage = `usage: %s <command>

commands:

  serve                                            Start FASS service.
  token <mail-file>                                Generate a token for each mail address and produces a 'mapping.json' file.
  course <identifier> <mapping-file>               Generate a course, adding the tokens from the given mapping file.
  distribute <course-identifier> <mapping-file>    Distributes the generated tokens via mail.
`

	fmt.Fprintf(os.Stderr, usage, os.Args[0])
	flag.PrintDefaults()
}

func generateTokenMapping(mailFilepath string) {
	const mappingPath = "mapping.json"

	mailFile, err := os.Open(mailFilepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer mailFile.Close()

	mapping := fass.NewTokenMapping(mailFile)

	mappingFile, err := os.Create(mappingPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer mappingFile.Close()

	mappingJSON, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	mappingFile.Write(mappingJSON)
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
	}

	for token := range mapping {
		course.Users = append(course.Users, token)
	}

	fass.StoreCourse(course)
}

func distributeTokens(courseIdentifier string, mappingFilepath string) {
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
			err := fass.DistributeToken(user, addr, course);
			if err != nil {
				fmt.Fprintln(os.Stderr, addr, err.Error())
			}
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "serve":
		fass.Serve("localhost:8080")
	case "token":
		generateTokenMapping(os.Args[2])
	case "course":
		generateCourse(os.Args[2], os.Args[3])
	case "distribute":
		distributeTokens(os.Args[2], os.Args[3])
	case "help":
		printUsage()
		os.Exit(0)
	default:
		printUsage()
		os.Exit(2)
	}
}
