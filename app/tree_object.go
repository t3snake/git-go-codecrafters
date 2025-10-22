package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

type TreeObjectItem struct {
	mode     string
	name     string
	sha_byte [20]byte
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

func WriteTree(path string) (string, [20]byte, error) {
	tree_items := make([]TreeObjectItem, 0)
	entries, _ := os.ReadDir(path)

	for _, entry := range entries {
		file_or_dir_path := fmt.Sprintf("%s/%s", path, entry.Name())

		if entry.IsDir() {
			// Directory case: Recursively do WriteFile for tree object within
			if entry.Name() == ".git" {
				continue
			}

			_, sha_byte, err := WriteTree(file_or_dir_path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				continue
			}

			tree_items = append(tree_items, TreeObjectItem{
				name:     entry.Name(),
				sha_byte: sha_byte,
				mode:     DirectoryMode[1:],
			})
			continue
		}
		// File case: Do hash-object for files
		var mode string

		_, sha_byte, err := HashObject(file_or_dir_path, true)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		file_info, err := entry.Info()
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		if file_info.Mode()&0111 != 0 {
			// executable file case
			mode = ExecutableFileMode
		} else if file_info.Mode().IsRegular() {
			mode = RegularFileMode
		} else if file_info.Mode()&fs.ModeSymlink != 0 {
			mode = SymbolicLinkMode
		}

		tree_items = append(tree_items, TreeObjectItem{
			name:     entry.Name(),
			sha_byte: sha_byte,
			mode:     mode,
		})
	}

	// Generate tree object content
	content := ""
	for _, tree_item := range tree_items {
		output_entry := fmt.Sprintf("%s %s%s%s", tree_item.mode, tree_item.name, NullChar, tree_item.sha_byte)
		content = fmt.Sprintf("%s%s", content, output_entry)
	}
	header := fmt.Sprintf("tree %d%s", len(content), NullChar)
	content = fmt.Sprintf("%s%s", header, content)

	// return sha of that content
	sha_hex, sha_bytes := writeContentAndGetShaHexAndBytes(content)

	// write zlib compressed content into .git folder
	err := compressAndWriteGitObject(content, sha_hex)
	return sha_hex, sha_bytes, err
}
