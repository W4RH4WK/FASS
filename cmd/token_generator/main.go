package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/W4RH4WK/FASS/pkg/fass"
)

func printUsage() {
	fmt.Fprintf(os.Stderr, "usage: %s <mail-file>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	mailFile, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	defer mailFile.Close()

	mapping, err := fass.NewTokenMapping(mailFile)
	if err != nil {
		log.Fatal(err)
	}

	var users []fass.Token
	for token := range mapping {
		users = append(users, token)
	}

	course := fass.Course{
		Identifier: "703000",
		Name:       "Test Course",
		URL:        "https://example.org",
		Users:      users,
	}

	result, _ := json.MarshalIndent(course, "", "  ")
	fmt.Println(string(result))
}
