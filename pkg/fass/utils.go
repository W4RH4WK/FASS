package fass

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

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

func isZIP(file io.Reader) bool {
	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return false
	}
	return http.DetectContentType(buffer) == "application/zip"
}
