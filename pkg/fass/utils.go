package fass

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

func marshalToFile(filepath string, v interface{}) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	vJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	file.Write(vJSON)
	return err
}

func unmarshalFromFile(filepath string, v interface{}) error {
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	return json.Unmarshal(content, v)
}

func isZIP(file *bufio.Reader) bool {
	header, _ := file.Peek(512)
	return http.DetectContentType(header) == "application/zip"
}
