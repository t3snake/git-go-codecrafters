package main

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
)

// returns path of the directory and full path of file respectively, given the SHA1 hash
func getDirAndFilePathFromHash(sha_hash string) (string, string) {
	directory_path := fmt.Sprintf(".git/objects/%s", sha_hash[:2])
	object_path := fmt.Sprintf("%s/%s", directory_path, sha_hash[2:])

	return directory_path, object_path
}

// returns 40 character hex SHA and 20 byte SHA given the content
func getShaHexAndBytesForContent(content string) (string, [20]byte) {
	sha_bytes := sha1.Sum([]byte(content))
	sha_hex := fmt.Sprintf("%x", sha_bytes)
	return sha_hex, sha_bytes
}

// writes content after zlib compression into .git/objects directory in directory sha[:2] and filename sha[2:]
func compressAndWriteGitObject(content string, sha_hash string) error {
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

	_, err = w.Write([]byte(content))
	return err
}
