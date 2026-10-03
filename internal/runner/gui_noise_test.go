package runner

import "testing"

// Qt announces its own VNC server on stderr at startup. That heading is where
// a student's traceback appears, and a panel that always has something in it
// is a panel they stop reading.
func TestQtStartupNoiseIsNotShownAsAnError(t *testing.T) {
	if got := dropQtNoise("QVncServer created on port 5900\n"); got != "" {
		t.Errorf("Qt's own notice must not reach the student, got %q", got)
	}

	real := "Traceback (most recent call last):\n  File \"main.py\", line 2\nNameError: name 'x' is not defined\n"
	mixed := "QVncServer created on port 5900\n" + real
	got := dropQtNoise(mixed)
	if got == "" {
		t.Fatal("a real traceback must survive")
	}
	if want := real; got != want {
		t.Errorf("traceback changed:\n got %q\nwant %q", got, want)
	}
}
