package main

import (
	"strings"
	"testing"
)

// The live diagnostic authors append as "(сейчас $(…))" loses its opening paren
// when studentFail cuts at the word, so the student used to be shown a message
// ending in a bracket that opens nothing.
func TestStudentFailDoesNotLeaveAStrayBracket(t *testing.T) {
	bad := "sshd игнорирует ключи при слишком открытых правах: chmod 700 ~/.ssh " +
		"(сейчас $(stat -c '%a' /root/.ssh) и $(stat -c '%a' /root/.ssh/authorized_keys))"
	got := studentFail("права приведены к безопасным", bad)

	if strings.Contains(got, "chmod 700") {
		t.Errorf("the solution leaked into the message: %q", got)
	}
	if open, closed := strings.Count(got, "("), strings.Count(got, ")"); closed > open {
		t.Errorf("unbalanced brackets (%d open, %d closed): %q", open, closed, got)
	}
	if !strings.Contains(got, "stat -c") {
		t.Errorf("the diagnostic was dropped: %q", got)
	}

	// A diagnostic that is balanced on its own must be left alone, parens included.
	keep := studentFail("ок", "Сейчас: $(ls -l) (точно так)")
	if !strings.HasSuffix(keep, "(точно так)") {
		t.Errorf("a balanced tail was trimmed: %q", keep)
	}
	// And without a diagnostic nothing is appended.
	if got := studentFail("ок", "mkdir /root/x"); got != "не выполнено: ок" {
		t.Errorf("plain solution text = %q", got)
	}
}
