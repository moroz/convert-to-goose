package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func resolveMigrationsDirectory(dir string) (string, error) {
	filename, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	if stat, err := os.Stat(filename); err != nil || !stat.IsDir() {
		return "", errors.New("directory not readable")
	}
	return filename, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: convert-to-goose DIRECTORY")
		os.Exit(2)
	}
	dir := os.Args[1]
	dir, err := resolveMigrationsDirectory(dir)
	if err != nil {
		log.Fatal(err)
	}
}
