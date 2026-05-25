package collector

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

const fixturePath = "../testdata/proc/fs/lustre"

// lnetctlNoop is a binary that doesn't exist; LNet errors are tolerated (fail open).
const lnetctlNoop = "/nonexistent/lnetctl"

func TestCollectorOSCMetrics(t *testing.T) {
	c := New(fixturePath, lnetctlNoop)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	// Gather and check a subset of expected metrics
	mfs, err := reg.Gather()
	if err != nil {
		// ContinueOnError — partial results are OK; log but don't fail on lnet errors
		t.Logf("Gather errors (expected for missing lnetctl): %v", err)
	}

	found := map[string]bool{}
	for _, mf := range mfs {
		found[mf.GetName()] = true
	}

	required := []string{
		"lustre_osc_state",
		"lustre_osc_grant_bytes",
		"lustre_osc_cached_mb",
		"lustre_osc_unstable_pages",
		"lustre_osc_read_rpcs_in_flight",
		"lustre_osc_write_rpcs_in_flight",
		"lustre_mdc_state",
		"lustre_mdc_ops_total",
		"lustre_mdc_modify_rpcs_in_flight",
	}
	for _, name := range required {
		if !found[name] {
			t.Errorf("metric %q not found in output", name)
		}
	}
}

func TestCollectorOSCState(t *testing.T) {
	c := New(fixturePath, lnetctlNoop)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	expected := `
# HELP lustre_osc_state OSC connection state (1=FULL, 0=other)
# TYPE lustre_osc_state gauge
lustre_osc_state{ost="pfs0-OST0000",server="10.200.8.17@tcp",target="pfs0-OST0000_UUID"} 1
lustre_osc_state{ost="pfs0-OST0001",server="10.200.8.19@tcp",target="pfs0-OST0001_UUID"} 0
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "lustre_osc_state"); err != nil {
		t.Error(err)
	}
}

func TestCollectorMDCOps(t *testing.T) {
	c := New(fixturePath, lnetctlNoop)
	reg := prometheus.NewRegistry()
	reg.MustRegister(c)

	// Verify a couple of MDC op counters exist with correct values
	expected := `
# HELP lustre_mdc_ops_total MDC metadata operation count
# TYPE lustre_mdc_ops_total counter
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="close",target="pfs0-MDT0000_UUID"} 1638616
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="create",target="pfs0-MDT0000_UUID"} 12345
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="getattr",target="pfs0-MDT0000_UUID"} 5678901
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="intent_lock",target="pfs0-MDT0000_UUID"} 24227828
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="link",target="pfs0-MDT0000_UUID"} 0
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="rename",target="pfs0-MDT0000_UUID"} 8765
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="setattr",target="pfs0-MDT0000_UUID"} 234567
lustre_mdc_ops_total{mdc="pfs0-MDT0000",operation="unlink",target="pfs0-MDT0000_UUID"} 56789
`
	if err := testutil.GatherAndCompare(reg, strings.NewReader(expected), "lustre_mdc_ops_total"); err != nil {
		t.Error(err)
	}
}
