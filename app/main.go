package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
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
			}
		}

		headFileContents := []byte("ref: refs/heads/main\n")
		if err := os.WriteFile(".git/HEAD", headFileContents, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %s\n", err)
		}

		fmt.Println("Initialized git directory")

	case "cat-file":
		// git cat-file -p <hash>
		content, err := CatFile(os.Args[3])
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
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

		hash, err := HashObject(filename, write_flag)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		fmt.Println(hash)

	case "ls-tree":
		// git ls-tree [--name-only] <tree_sha>
		// implement read tree object
		var tree_hash string

		is_name_only := os.Args[2] == "--name-only"

		if is_name_only {
			tree_hash = os.Args[3]
		} else {
			tree_hash = os.Args[2]
		}

		content, err := LsTree(tree_hash, is_name_only)
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		fmt.Println(content)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

// CatFile reads and returns contents of a blob object stored in .git given its hash.
func CatFile(hash string) (string, error) {
	_, file_path := getDirAndFilePathFromHash(hash)

	file, err := os.ReadFile(file_path)
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
	if strings.Contains(blob, NullChar) {
		content := strings.Split(blob, NullChar)[1]
		return content, nil
	} else {
		return "", fmt.Errorf("invalid blob: doesnt contain null byte")
	}

}

// HashObject returns hash of given file as a blob. If write_flag is true, writes the blob into .git
func HashObject(filename string, write_flag bool) (string, error) {
	file_content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	blob := generateBlobObject(string(file_content))

	sha := fmt.Sprintf("%x", sha1.Sum([]byte(blob)))

	if write_flag {
		err = writeBlobObject(blob, sha)
		if err != nil {
			return "", err
		}
	}

	return sha, nil
}

// generateBlobObject generates blob string and returns it
func generateBlobObject(content string) string {
	content_length := len(content)

	return fmt.Sprintf("blob %d%s%s", content_length, NullChar, content)
}

// writeBlobObject writes a blob into .git directory
func writeBlobObject(blob_content string, sha_hash string) error {
	directory_path, object_path := getDirAndFilePathFromHash(sha_hash)

	os.Mkdir(directory_path, 0755)

	// get new os.File which implements io.Writer
	file, err := os.OpenFile(object_path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// compress blob with zlib
	w := zlib.NewWriter(file)
	defer w.Close()

	_, err = w.Write([]byte(blob_content))

	// err := os.WriteFile(object_path, b.Bytes(), 0644)
	return err
}

// returns path of the directory and full path of file respectively, given the SHA1 hash
func getDirAndFilePathFromHash(sha_hash string) (string, string) {
	directory_path := fmt.Sprintf(".git/objects/%s", sha_hash[:2])
	object_path := fmt.Sprintf("%s/%s", directory_path, sha_hash[2:])

	return directory_path, object_path
}

// returns contents of tree object, given tree object SHA1 hash
func LsTree(sha_hash string, is_name_only bool) (string, error) {
	_, file_path := getDirAndFilePathFromHash(sha_hash)

	file, err := os.ReadFile(file_path)
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

	decompressed_tree_obj := b.String()

	return parseTreeObject(decompressed_tree_obj, is_name_only)
}

// parses decompressed content of tree object and returns output in ls-tree specified format
func parseTreeObject(contents string, is_name_only bool) (string, error) {
	// Input Format:
	// tree <size>\0
	// <mode> <name>\0<20_byte_sha>
	// <mode> <name>\0<20_byte_sha> ...
	output := ""
	segments := strings.Split(contents, NullChar)

	var mode, file_name, sha_hash string
	// TODO may need validation to check if content is valid
	for i := range segments {
		if i == 0 {
			// ignore segments[0] for now, may need validation later
			continue
		} else if i == 1 {
			// No SHA on the first segment
			sub_segments := strings.Split(segments[i], " ")
			mode = sub_segments[0]
			file_name = sub_segments[1]
		} else if i == len(segments)-1 {
			sha_hash = segments[i][:20]
			// write last entry to output now before moving to next entry
			if is_name_only {
				output = fmt.Sprintf("%s%s\n", output, file_name)
			} else {
				var tree_or_blob string
				if mode == DirectoryMode {
					tree_or_blob = "tree"
				} else {
					tree_or_blob = "blob"
				}
				output = fmt.Sprintf("%s%s %s %s    %s", output, mode, tree_or_blob, sha_hash, file_name)
			}
		} else {
			sha_hash = segments[i][:20]

			// write the entry to output now before moving to next entry
			if is_name_only {
				output = fmt.Sprintf("%s%s\n", output, file_name)
			} else {
				var tree_or_blob string
				if mode == DirectoryMode {
					tree_or_blob = "tree"
				} else {
					tree_or_blob = "blob"
				}
				output = fmt.Sprintf("%s%s %s %s    %s", output, mode, tree_or_blob, sha_hash, file_name)
			}

			// parse rest of segment, new entry
			sub_segments := strings.Split(segments[i][20:], " ")
			mode = sub_segments[0]
			file_name = sub_segments[1]
		}
	}

	return output, nil
}
