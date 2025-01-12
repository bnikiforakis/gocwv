package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

type URLList struct {
	URLs []string `json:"urls"`
}

func ReadURLsFromFile(filename string) ([]string, error) {
	//Debug Step: Print the file that it reads from
	fmt.Println("Reading URLs from file:", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return nil, err
	}

	//Debug Step: Print the URLS to verify that it works properly
	fmt.Println("Read file contents:", string(data))

	var urlList URLList
	err = json.Unmarshal(data, &urlList)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	return urlList.URLs, nil
}
