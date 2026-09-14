package lab

import (
	"fmt"
	"strings"
)

// Ref types in a commit graph.
const (
	RefHead   = "head"
	RefBranch = "branch"
	RefRemote = "remote"
	RefTag    = "tag"
)

// Commit is one node of the git graph.
type Commit struct {
	Hash    string
	Parents []string
	Refs    []Ref
	Subject string
}

// Ref is a branch, tag, remote branch or HEAD pointing at a commit.
type Ref struct {
	Name string
	Type string
}

const graphLimit = 60

// GitLogCommand prints the newest commits of repo in the format ParseGitLog reads.
func GitLogCommand(repo string) string {
	return fmt.Sprintf(`git -C '%s' log --all --date-order --decorate=full --pretty=tformat:'%%h%%x1f%%p%%x1f%%D%%x1f%%s' 2>/dev/null | head -%d`,
		strings.ReplaceAll(repo, "'", ""), graphLimit)
}

// ParseGitLog parses GitLogCommand output; unparsable lines (e.g. errors) are skipped.
func ParseGitLog(out string) []Commit {
	commits := []Commit{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Split(strings.TrimRight(line, "\r"), "\x1f")
		if len(f) != 4 || f[0] == "" {
			continue
		}
		commits = append(commits, Commit{Hash: f[0], Parents: strings.Fields(f[1]), Refs: parseRefs(f[2]), Subject: f[3]})
	}
	return commits
}

func parseRefs(decoration string) []Ref {
	refs := []Ref{}
	for _, d := range strings.Split(decoration, ", ") {
		d = strings.TrimSpace(d)
		switch {
		case d == "":
		case d == "HEAD":
			refs = append(refs, Ref{Name: "HEAD", Type: RefHead})
		case strings.HasPrefix(d, "HEAD -> "):
			refs = append(refs, Ref{Name: "HEAD", Type: RefHead}, fullRef(strings.TrimPrefix(d, "HEAD -> ")))
		case strings.HasPrefix(d, "tag: "):
			refs = append(refs, fullRef(strings.TrimPrefix(d, "tag: ")))
		default:
			refs = append(refs, fullRef(d))
		}
	}
	return refs
}

func fullRef(name string) Ref {
	switch {
	case strings.HasPrefix(name, "refs/heads/"):
		return Ref{Name: strings.TrimPrefix(name, "refs/heads/"), Type: RefBranch}
	case strings.HasPrefix(name, "refs/remotes/"):
		return Ref{Name: strings.TrimPrefix(name, "refs/remotes/"), Type: RefRemote}
	case strings.HasPrefix(name, "refs/tags/"):
		return Ref{Name: strings.TrimPrefix(name, "refs/tags/"), Type: RefTag}
	default:
		return Ref{Name: name, Type: RefBranch}
	}
}
