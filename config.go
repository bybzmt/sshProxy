package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
)

func loadFile(file string) (string, error) {
	f, err := openFile(&file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func openFile(file *string) (*os.File, error) {
	f, err := os.Open(*file)
	if err == nil {
		log.Println("openFile", *file)
		return f, nil
	}

	// If no config file in current directory, try search it in the binary directory
	// Note there's no portable way to detect the binary directory.
	binDir := path.Dir(os.Args[0])
	if binDir != "" && binDir != "." {
		*file = path.Join(binDir, *file)
		f, err = os.Open(*file)
		if err == nil {
			log.Println("openFile", *file)
		}
	}
	return f, err
}
