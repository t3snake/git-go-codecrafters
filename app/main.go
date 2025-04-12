package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"strings"
)

// Usage: your_program.sh <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// Uncomment this block to pass the first stage!

		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		content, err := uncompressZlib(os.Args[3])
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		fmt.Print(content)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func uncompressZlib(hash string) (string, error) {
	file_name := ".git/objects/" + hash[0:2] + "/" + hash[2:]

	file, err := os.ReadFile(file_name)
	if err != nil {
		return "", err
	}

	buf := bytes.NewReader(file)

	r, err := zlib.NewReader(buf)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var b bytes.Buffer
	_, err = b.ReadFrom(r)
	if err != nil {
		return "", fmt.Errorf("error reading blob file: %s", err.Error())
	}

	blob := b.String()
	if strings.Contains(blob, string(rune(0))) {
		content := strings.Split(blob, string(rune(0)))[1]
		return content, nil
	} else {
		return "", fmt.Errorf("invalid blob: doesnt contain null byte")
	}

}
