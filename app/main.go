package main

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Fprintf(os.Stderr, "Logs from your program will appear here!\n")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: mygit <command> [<args>...]\n")
		os.Exit(1)
	}

	switch command := os.Args[1]; command {
	case "init":
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
		content, err := CatFile(os.Args[3])
		if err != nil {
			fmt.Fprint(os.Stderr, err.Error())
		}
		fmt.Print(content)

	case "hash-object":
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

	default:
		fmt.Fprintf(os.Stderr, "Unknown command %s\n", command)
		os.Exit(1)
	}
}

// CatFile reads and returns contents of a blob object stored in .git given its hash.
func CatFile(hash string) (string, error) {
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
	return fmt.Sprintf("blob %d%s%s", content_length, string(rune(0)), content)
}

// writeBlobObject writes a blob into .git directory
func writeBlobObject(blob_content string, sha_hash string) error {
	// ".git/objects/" + hash[0:2] + "/" + hash[2:]
	object_path := fmt.Sprintf(".git/objects/%s/%s", sha_hash[:2], sha_hash[2:])

	// get new os.File which implements io.Writer
	file, err := os.OpenFile(object_path, os.O_CREATE, 0644)
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
