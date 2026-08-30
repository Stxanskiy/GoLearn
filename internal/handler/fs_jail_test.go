package handler

import "testing"

func TestJailPath(t *testing.T) {
	ok := map[string]string{
		"":                      "/root",
		"/root":                 "/root",
		"/root/project":         "/root/project",
		"/root/project/app.yml": "/root/project/app.yml",
		"project/app.yml":       "/root/project/app.yml",
		"/root/../root/x":       "/root/x",
	}
	for in, want := range ok {
		got, jok := jailPath(in)
		if !jok || got != want {
			t.Errorf("jailPath(%q) = %q,%v; want %q,true", in, got, jok, want)
		}
	}
	bad := []string{"/etc/passwd", "/root/../etc/shadow", "/", "/rootx", "../../etc", "/root/../.."}
	for _, in := range bad {
		if got, jok := jailPath(in); jok {
			t.Errorf("jailPath(%q) allowed escape -> %q", in, got)
		}
	}
}
