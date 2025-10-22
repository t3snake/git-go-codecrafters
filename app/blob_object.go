package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"os"
	"strings"
)

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
func HashObject(filename string, write_flag bool) (string, [20]byte, error) {
	empty_bytes := [20]byte{}
	file_content, err := os.ReadFile(filename)
	if err != nil {
		return "", empty_bytes, err
	}

	blob := generateBlobObject(string(file_content))

	sha_hex, sha_byte := writeContentAndGetShaHexAndBytes(blob)

	if write_flag {
		err = compressAndWriteGitObject(blob, sha_hex)
		if err != nil {
			return "", empty_bytes, err
		}
	}

	return sha_hex, sha_byte, nil
}

// generateBlobObject generates blob string and returns it
func generateBlobObject(content string) string {
	content_length := len(content)

	return fmt.Sprintf("blob %d%s%s", content_length, NullChar, content)
}
