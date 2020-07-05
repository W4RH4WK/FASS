package fass

import (
	"encoding/json"
	"io/ioutil"
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
