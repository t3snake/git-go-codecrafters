package main

import (
	"fmt"
	"time"
)

// writes to .git/objects, a commit object using tree hash and optional parent hash and commit message.
func CommitTree(tree_hash, parent_hash, commit_message string) (string, error) {
	// assumes currently that we always get a single parent hash and one line in commit message
	tree_line := fmt.Sprintf("tree %s\n", tree_hash)
	parent_line := fmt.Sprintf("parent %s\n", parent_hash)

	time_now := time.Now()
	timestamp := time_now.Unix()
	timezone := time_now.Local().Format("-0700")

	// currently hardcoded name and email
	author_line := fmt.Sprintf("author %s <%s> %d %s\n", "t3snake", "t3snake@gmail.com", timestamp, timezone)
	committer_line := fmt.Sprintf("committer %s <%s> %d %s\n", "t3snake", "t3snake@gmail.com", timestamp, timezone)

	body := fmt.Sprintf("%s%s%s%s\n%s\n", tree_line, parent_line, author_line, committer_line, commit_message)
	header := fmt.Sprintf("commit %d%s", len(body), NullChar)
	content := fmt.Sprintf("%s%s", header, body)

	sha_hex, _ := getShaHexAndBytesForContent(content)
	err := compressAndWriteGitObject(content, sha_hex)
	return sha_hex, err
}
