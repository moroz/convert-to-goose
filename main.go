package main

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

func resolveGitRoot(dir string) (string, error) {
	parent := dir
	for {
		if parent == "/" {
			return "", errors.New("the specified path is not inside a Git repository")
		}
		gitFolderPath := filepath.Join(parent, ".git")
		if stat, err := os.Stat(gitFolderPath); err == nil && stat.IsDir() {
			return parent, nil
		}
		parent = filepath.Dir(parent)
	}
}

func checkForUncommittedChanges(gitRoot string) (bool, error) {
	stdout := bytes.NewBuffer([]byte{})
	cmd := exec.Command("git", "status", "--porcelain=v1")
	cmd.Dir = gitRoot
	cmd.Stdout = stdout
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	return stdout.Len() == 0, nil
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

	gitRoot, err := resolveGitRoot(dir)
	if err != nil {
		log.Fatal(err)
	}

	clean, err := checkForUncommittedChanges(gitRoot)
	if err != nil {
		log.Fatal(err)
	}
	if !clean {
		log.Fatal("The working tree is not clean. Please commit your changes before continuing.")
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	targets := map[string]struct{}{}

	replacer := regexp.MustCompile(".(down|up).sql$")

	for _, file := range files {
		name := file.Name()
		if !replacer.MatchString(name) {
			continue
		}
		target := replacer.ReplaceAllString(name, "")
		targets[target] = struct{}{}
	}

	for target := range targets {
		upPath := filepath.Join(dir, target+".up.sql")
		up, err := os.ReadFile(upPath)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove(upPath)
		if err != nil {
			log.Fatal(err)
		}

		downPath := filepath.Join(dir, target+".down.sql")
		down, err := os.ReadFile(downPath)
		if err != nil {
			log.Fatal(err)
		}
		err = os.Remove(downPath)
		if err != nil {
			log.Fatal(err)
		}

		content := "-- +goose Up\n-- +goose StatementBegin\n"
		content += strings.TrimSpace(string(up))
		content += "\n-- +goose StatementEnd"

		if len(down) != 0 {
			content += "\n\n-- +goose Down\n-- +goose StatementBegin\n"
			content += strings.TrimSpace(string(down))
			content += "\n-- +goose StatementEnd"
		}

		targetPath := filepath.Join(dir, target+".sql")
		err = os.WriteFile(targetPath, []byte(content), 0o664)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Printf("Written %d migrations!\n", len(targets))
}
