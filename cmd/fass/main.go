package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/W4RH4WK/FASS/pkg/fass"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s <command>\n", os.Args[0])
	flag.PrintDefaults()
}

func generateTokenMapping(mailFilepath string) {
	const mappingPath = "mapping.txt"

	mailFile, err := os.Open(mailFilepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer mailFile.Close()

	mapping, err := fass.NewTokenMapping(mailFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

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

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "token":
		generateTokenMapping(os.Args[2])
	case "serve":
		fass.Serve("localhost:8080")
	case "help":
		printUsage()
		os.Exit(0)
	default:
		printUsage()
		os.Exit(2)
	}

	// flag.Usage = printUsage
	// flag.Parse()

	// if flag.NArg() != 1 {
	// 	flag.Usage()
	// 	os.Exit(2)
	// }

	// mailFile, err := os.Open(flag.Args()[0])
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer mailFile.Close()

	// mapping, err := fass.NewTokenMapping(mailFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// var users []fass.Token
	// for token := range mapping {
	// 	users = append(users, token)
	// }

	// course := fass.Course{
	// 	Identifier: "703000",
	// 	Name:       "Test Course",
	// 	URL:        "https://example.org",
	// 	Users:      users,
	// }

	// result, _ := json.MarshalIndent(course, "", "  ")
	// fmt.Println(string(result))

	// course, err := fass.LoadCourses(".")
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Println(course)
}
