package lab

import (
	"errors"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/backendraz/golearn/internal/model"
)

func TestJailPath(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"", "/root", true},
		{"project/app.yml", "/root/project/app.yml", true},
		{"/root/../root/x", "/root/x", true},
		{"/root/../etc/passwd", "", false},
		{"../etc", "", false},
		{"/rootkit", "", false},
	}
	for _, c := range cases {
		got, ok := JailPath(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("JailPath(%q) = %q,%v want %q,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestInjectBase(t *testing.T) {
	if got := string(InjectBase([]byte("<html><HEAD><title>x</title>"), "/p/")); !strings.Contains(got, `<HEAD><base href="/p/"><title>`) {
		t.Errorf("head: %s", got)
	}
	if got := string(InjectBase([]byte("plain"), "/p/")); got != `<base href="/p/">plain` {
		t.Errorf("no head: %s", got)
	}
	if !strings.Contains(PreviewErrorPage(8080, errors.New("preview: curl-missing")), "нет curl") {
		t.Error("curl-missing reason")
	}
}

func TestIsGitLab(t *testing.T) {
	if !IsGitLab(model.Module{Category: "git"}) || !IsGitLab(model.Module{Slug: "git-basics"}) || IsGitLab(model.Module{Slug: "docker"}) {
		t.Error("IsGitLab")
	}
}

func TestParseGitLog(t *testing.T) {
	out := "fatal: not a git repository\n" +
		"f62e491\x1f003c910\x1fHEAD -> refs/heads/main\x1fsecond\n" +
		"003c910\x1f\x1ftag: refs/tags/v1, refs/remotes/origin/main, refs/heads/feature/x\x1ffirst | msg\n" +
		"a1\x1fb2 c3\x1fHEAD\x1fmerge\n"
	got := ParseGitLog(out)
	want := []Commit{
		{Hash: "f62e491", Parents: []string{"003c910"}, Refs: []Ref{{"HEAD", RefHead}, {"main", RefBranch}}, Subject: "second"},
		{Hash: "003c910", Parents: []string{}, Refs: []Ref{{"v1", RefTag}, {"origin/main", RefRemote}, {"feature/x", RefBranch}}, Subject: "first | msg"},
		{Hash: "a1", Parents: []string{"b2", "c3"}, Refs: []Ref{{"HEAD", RefHead}}, Subject: "merge"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseGitLog =\n%+v\nwant\n%+v", got, want)
	}
	if got := ParseGitLog(""); got == nil || len(got) != 0 {
		t.Errorf("empty output = %#v", got)
	}
}

func TestGitLogCommandAgainstRealRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=a", "-c", "user.email=a@b"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	run("commit", "-q", "--allow-empty", "-m", "one")
	run("branch", "feature/x")
	run("commit", "-q", "--allow-empty", "-m", "two")
	out, err := exec.Command("bash", "-c", GitLogCommand(dir)).Output()
	if err != nil {
		t.Fatal(err)
	}
	commits := ParseGitLog(string(out))
	if len(commits) != 2 || commits[0].Subject != "two" || len(commits[1].Refs) != 1 || commits[1].Refs[0] != (Ref{"feature/x", RefBranch}) {
		t.Fatalf("commits = %+v", commits)
	}
}
