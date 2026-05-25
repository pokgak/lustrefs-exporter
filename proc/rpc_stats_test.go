package proc

import "testing"

func TestParseRPCStatsHeader_OSC(t *testing.T) {
	h, err := ParseRPCStatsHeader("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/rpc_stats")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		got  int64
		want int64
	}{
		{"ReadRPCsInFlight", h.ReadRPCsInFlight, 3},
		{"WriteRPCsInFlight", h.WriteRPCsInFlight, 5},
		{"DIORPCsInFlight", h.DIORPCsInFlight, 1},
		{"PendingWritePages", h.PendingWritePages, 128},
		{"PendingReadPages", h.PendingReadPages, 64},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}

func TestParseRPCStatsHeader_MDC(t *testing.T) {
	h, err := ParseRPCStatsHeader("../testdata/proc/fs/lustre/mdc/pfs0-MDT0000-mdc-ff331f65467c7800/rpc_stats")
	if err != nil {
		t.Fatal(err)
	}
	if h.WriteRPCsInFlight != 2 {
		t.Errorf("WriteRPCsInFlight = %d, want 2", h.WriteRPCsInFlight)
	}
}
