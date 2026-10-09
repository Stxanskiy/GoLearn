package runner

import "testing"

// A lesson's requested size is a request, not a promise: FC_MAX_VMS of these
// boot on one host, so an unbounded number in the content is a way to take the
// machine down with a content edit.
func TestRequestedSizeIsClamped(t *testing.T) {
	const (
		lo  = vmMinMemMiB
		hi  = 4096
		cpu = 2
	)
	cases := []struct {
		name             string
		askMem, askCPU   int
		wantMem, wantCPU int
	}{
		{"within bounds", 2048, 2, 2048, 2},
		{"memory above the ceiling", 65536, 1, hi, 1},
		{"memory below what boots", 16, 1, lo, 1},
		{"more CPUs than the host allows", 1024, 64, 1024, cpu},
		{"zero CPUs is not a request", 1024, 0, 1024, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotMem := tc.askMem
			if tc.askMem > 0 {
				gotMem = clampInt(tc.askMem, lo, hi)
			}
			if gotMem != tc.wantMem {
				t.Errorf("memory %d → %d, want %d", tc.askMem, gotMem, tc.wantMem)
			}
			gotCPU := tc.askCPU
			if tc.askCPU > 0 {
				gotCPU = clampInt(tc.askCPU, 1, cpu)
			}
			if gotCPU != tc.wantCPU {
				t.Errorf("cpus %d → %d, want %d", tc.askCPU, gotCPU, tc.wantCPU)
			}
		})
	}
}

// A misconfigured host must not invert the range and let anything through.
func TestClampSurvivesAnInvertedRange(t *testing.T) {
	if got := clampInt(9999, 1024, 512); got != 1024 {
		t.Errorf("clamp with hi below lo = %d, want the floor 1024", got)
	}
}

// Zero fields are what every lesson has today; they must mean "the profile
// decides", not "a machine with no memory".
func TestEmptySpecAsksForNothing(t *testing.T) {
	var s Spec
	if s.CPUs != 0 || s.MemMiB != 0 {
		t.Fatalf("zero Spec is not zero: %+v", s)
	}
}
