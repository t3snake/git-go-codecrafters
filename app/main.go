package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
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
		fmt.Print(uncompressZlib(os.Args[3]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

func uncompressZlib(data string) string {
	var in bytes.Buffer
	var out []byte

	in.WriteString(data)
	r, err := zlib.NewReader(&in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decompressing zlib")
	}

	r.Read(out)
	r.Close()

	return string(out)
}
