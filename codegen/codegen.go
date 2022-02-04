package codegen

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/shpota/goxygen/static"
)

type generator struct {
	projectName string
	techStack   []string
}

func Generate(projectName string, techStack []string) {
	g := generator{projectName, techStack}

	// This code required to be able to run docker container so we could export generated files
	if os.Getenv("GOXYGEN_DOCKER") == "true" {
		g.projectName = "generated/" + g.projectName
	}

	g.generate()
}

func (g generator) generate() {
	fmt.Println("Generating", g.projectName)
	for path, srcText := range static.Sources() {
		srcText = strings.Replace(srcText, "project-name", g.projectName, -1)
		binary := []byte(srcText)
		g.processFile(path, binary)
	}
	for path, binary := range static.Images() {
		g.processFile(path, binary)
	}
	err := g.initGitRepo()
	if err != nil {
		fmt.Println("Failed to setup a Git repository:", err)
	}
	fmt.Println("Generation completed.")
}

// Checks if a file with the given path has to be generated, creates
// a directory structure, and a file with the given content.
func (g generator) processFile(path string, content []byte) {
	if !g.needed(path) {
		return
	}
	for _, tech := range g.techStack {
		path = strings.Replace(path, tech+".", "", 1)
	}
	pathElements := strings.Split(path, "/")
	separator := string(os.PathSeparator)
	pathElements = append([]string{g.projectName}, pathElements...)
	pathFile := strings.Join(pathElements, separator)

	_ = os.MkdirAll(
		strings.Join(pathElements[:len(pathElements)-1], separator),
		os.ModePerm,
	)

	fmt.Println("creating: " + pathFile)
	err := ioutil.WriteFile(
		pathFile,
		content,
		0644,
	)
	if err != nil {
		log.Fatal(err)
	}
}

func (g generator) initGitRepo() error {
	fmt.Println("setting up Git repository")
	cmd := exec.Command("git", "init", "-b", "main", ".")
	cmd.Dir = g.projectName
	err := cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = g.projectName
	err = cmd.Run()
	if err != nil {
		return err
	}
	cmd = exec.Command("git", "commit", "-m", "Initial commit from Goxygen")
	cmd.Dir = g.projectName
	return cmd.Run()
}

// Checks if a path is a framework-specific path (starts
// with framework name). Returns true if a path is
// prefixed with the provided framework followed by dot
// or if a path has no prefix.
func (g generator) needed(path string) bool {
	if !hasTechPrefix(path) {
		return true
	}
	for _, tech := range g.techStack {
		if strings.HasPrefix(path, tech+".") {
			return true
		}
	}
	return false
}

func hasTechPrefix(path string) bool {
	for _, tech := range []string{
		"angular", "react", "vue", "mongo", "mysql", "postgres",
	} {
		if strings.HasPrefix(path, tech+".") {
			return true
		}
	}
	return false
}
