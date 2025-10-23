package main

import (
	"fmt"
	"os"
)

const (
	NullChar           = string(rune(0))
	RegularFileMode    = "100644"
	ExecutableFileMode = "100755"
	SymbolicLinkMode   = "120000"
	DirectoryMode      = "040000"
)

func main() {
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
		// git init
		for _, dir := range []string{".git", ".git/objects", ".git/refs"} {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating directory: %s\n", err)
				os.Exit(1)
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		// git cat-file -p <hash>
		content, err := CatFile(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Print(content)

	case "hash-object":
		// git hash-object [-w] <filename>
		var filename string
		write_flag := os.Args[2] == "-w"

		if write_flag {
			filename = os.Args[3]
		} else {
			filename = os.Args[2]
		}

		hash, _, err := HashObject(filename, write_flag)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(hash)

	case "ls-tree":
		// git ls-tree [--name-only] <tree_sha>
		var tree_hash string

		is_name_only := os.Args[2] == "--name-only"

		if is_name_only {
			tree_hash = os.Args[3]
		} else {
			tree_hash = os.Args[2]
		}

		content, err := LsTree(tree_hash, is_name_only)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Print(content) // newline already added within implementation

	case "write-tree":
		// git write-tree
		hash, _, err := WriteTree(".")
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(hash)

	case "commit-tree":
		// git commit-tree <tree_hash> -p <parent_hash> -m <message>
		parent_hash := ""
		commit_message := ""

		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Invalid number of arguments. Usage: git commit-tree <tree_hash> [options -m -p]")
			os.Exit(1)
		}
		tree_hash := os.Args[2]

		switch os.Args[3] {
		case "-p":
			if len(os.Args) < 5 {
				fmt.Fprintln(os.Stderr, "Missing parent hash after -p")
				os.Exit(1)
			}
			parent_hash = os.Args[4]
		case "-m":
			if len(os.Args) < 5 {
				fmt.Fprintln(os.Stderr, "Missing commit message after -m")
				os.Exit(1)
			}
			commit_message = os.Args[4]
		}

		switch os.Args[5] {
		case "-p":
			if len(os.Args) < 7 {
				fmt.Fprintf(os.Stderr, "Missing parent hash after -p")
				os.Exit(1)
			}
			parent_hash = os.Args[6]
		case "-m":
			if len(os.Args) < 7 {
				fmt.Fprintf(os.Stderr, "Missing commit message after -m")
				os.Exit(1)
			}
			commit_message = os.Args[6]
		}

		hash, err := CommitTree(tree_hash, parent_hash, commit_message)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		fmt.Println(hash)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}
