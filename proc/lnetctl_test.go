package proc

import (
	"os"
	"testing"
)

func TestParseLNetStats(t *testing.T) {
	data, err := os.ReadFile("../testdata/lnetctl_stats.yaml")
	if err != nil {
		t.Fatal(err)
	}
	stats, err := ParseLNetStats(data)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		got  int64
		want int64
	}{
		{"SendCount", stats.SendCount, 72420754},
		{"RecvCount", stats.RecvCount, 63612337},
		{"Errors", stats.Errors, 3},
		{"ResendCount", stats.ResendCount, 12},
		{"ResponseTimeoutCount", stats.ResponseTimeoutCount, 2},
		{"NetworkTimeoutCount", stats.NetworkTimeoutCount, 16},
	}
	for _, tt := range cases {
		if tt.got != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestParseLNetExport(t *testing.T) {
	data, err := os.ReadFile("../testdata/lnetctl_export.yaml")
	if err != nil {
		t.Fatal(err)
	}
	nis, err := ParseLNetExport(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(nis) != 2 {
		t.Fatalf("got %d NIs, want 2", len(nis))
	}

	up := nis[0]
	if up.NID != "10.20.0.8@tcp1345" {
		t.Errorf("NID = %q, want 10.20.0.8@tcp1345", up.NID)
	}
	if up.Status != "up" {
		t.Errorf("Status = %q, want up", up.Status)
	}
	if up.NetType != "tcp1345" {
		t.Errorf("NetType = %q, want tcp1345", up.NetType)
	}
	if up.Interface != "bond0" {
		t.Errorf("Interface = %q, want bond0", up.Interface)
	}

	down := nis[1]
	if down.Status != "down" {
		t.Errorf("Status = %q, want down", down.Status)
	}
}
