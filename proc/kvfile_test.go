package proc

import (
	"testing"
)

func TestParseKeyValueFile_OscStats(t *testing.T) {
	m, err := ParseKeyValueFile("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/osc_stats")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]int64{
		"lockless_write_bytes": 2048576,
		"lockless_read_bytes":  1024288,
	}
	for k, want := range cases {
		if got := m[k]; got != want {
			t.Errorf("%s = %d, want %d", k, got, want)
		}
	}
}

func TestParseKeyValueFile_CachedMB(t *testing.T) {
	m, err := ParseKeyValueFile("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/osc_cached_mb")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]int64{
		"used_mb":  128,
		"busy_cnt": 4,
		"reclaim":  256,
		"unevict":  0,
	}
	for k, want := range cases {
		if got := m[k]; got != want {
			t.Errorf("%s = %d, want %d", k, got, want)
		}
	}
}

func TestParseKeyValueFile_UnstableStats(t *testing.T) {
	m, err := ParseKeyValueFile("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/unstable_stats")
	if err != nil {
		t.Fatal(err)
	}
	if m["unstable_pages"] != 512 {
		t.Errorf("unstable_pages = %d, want 512", m["unstable_pages"])
	}
	if m["unstable_mb"] != 2 {
		t.Errorf("unstable_mb = %d, want 2", m["unstable_mb"])
	}
}

func TestReadSingleInt64(t *testing.T) {
	v, err := ReadSingleInt64("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/cur_grant_bytes")
	if err != nil {
		t.Fatal(err)
	}
	if v != 4194304 {
		t.Errorf("got %d, want 4194304", v)
	}
}
