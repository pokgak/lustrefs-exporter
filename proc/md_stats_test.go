package proc

import "testing"

func TestParseMDStats(t *testing.T) {
	ops, err := ParseMDStats("../testdata/proc/fs/lustre/mdc/pfs0-MDT0000-mdc-ff331f65467c7800/md_stats")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]int64{
		"close":       1638616,
		"create":      12345,
		"getattr":     5678901,
		"intent_lock": 24227828,
		"setattr":     234567,
		"unlink":      56789,
	}
	for op, want := range cases {
		if got := ops[op]; got != want {
			t.Errorf("op %s = %d, want %d", op, got, want)
		}
	}
	// snapshot_time line should not appear
	if _, ok := ops["snapshot_time"]; ok {
		t.Error("snapshot_time should not be in ops map")
	}
}
