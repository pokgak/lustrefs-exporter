package proc

import "testing"

func TestParseImport_Full(t *testing.T) {
	info, err := ParseImport("../testdata/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/import")
	if err != nil {
		t.Fatal(err)
	}
	if info.Target != "pfs0-OST0000_UUID" {
		t.Errorf("target = %q, want pfs0-OST0000_UUID", info.Target)
	}
	if info.State != "FULL" {
		t.Errorf("state = %q, want FULL", info.State)
	}
	if info.CurrentConnection != "10.200.8.17@tcp" {
		t.Errorf("connection = %q, want 10.200.8.17@tcp", info.CurrentConnection)
	}
}

func TestParseImport_Connecting(t *testing.T) {
	info, err := ParseImport("../testdata/proc/fs/lustre/osc/pfs0-OST0001-osc-ff331f65467c7801/import")
	if err != nil {
		t.Fatal(err)
	}
	if info.State != "CONNECTING" {
		t.Errorf("state = %q, want CONNECTING", info.State)
	}
}

func TestParseImport_MDC(t *testing.T) {
	info, err := ParseImport("../testdata/proc/fs/lustre/mdc/pfs0-MDT0000-mdc-ff331f65467c7800/import")
	if err != nil {
		t.Fatal(err)
	}
	if info.Target != "pfs0-MDT0000_UUID" {
		t.Errorf("target = %q, want pfs0-MDT0000_UUID", info.Target)
	}
	if info.State != "FULL" {
		t.Errorf("state = %q, want FULL", info.State)
	}
}
