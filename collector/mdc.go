package collector

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pokgak/lustrefs-exporter/proc"
)

type mdcCollector struct {
	lustrePath string

	state      *prometheus.Desc
	ops        *prometheus.Desc
	modifyRPCs *prometheus.Desc
}

var mdcLabels = []string{"target", "mdc"}

func newMDCCollector(lustrePath string) *mdcCollector {
	d := func(name, help string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc("lustre_"+name, help, labels, nil)
	}
	return &mdcCollector{
		lustrePath: lustrePath,
		state:      d("mdc_state", "MDC connection state (1=FULL, 0=other)", append(mdcLabels, "server")...),
		ops:        d("mdc_ops_total", "MDC metadata operation count", append(mdcLabels, "operation")...),
		modifyRPCs: d("mdc_modify_rpcs_in_flight", "MDC modify RPCs in flight", mdcLabels...),
	}
}

func (c *mdcCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.state
	ch <- c.ops
	ch <- c.modifyRPCs
}

func (c *mdcCollector) collect(ch chan<- prometheus.Metric) {
	dirs, err := filepath.Glob(filepath.Join(c.lustrePath, "mdc", "*"))
	if err != nil || len(dirs) == 0 {
		return
	}

	for _, dir := range dirs {
		name := filepath.Base(dir)
		mdc := mdcDirToMDC(name)
		if err := c.collectMDT(ch, dir, mdc); err != nil {
			logWarn(fmt.Sprintf("mdc %s", name), err)
		}
	}
}

// mdcDirToMDC extracts the MDT name from directory like pfs0-MDT0000-mdc-ff331f65467c7800.
func mdcDirToMDC(dir string) string {
	parts := strings.Split(dir, "-mdc-")
	if len(parts) >= 1 {
		return parts[0]
	}
	return dir
}

func (c *mdcCollector) collectMDT(ch chan<- prometheus.Metric, dir, mdc string) error {
	imp, err := proc.ParseImport(filepath.Join(dir, "import"))
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	stateVal := 0.0
	if imp.State == "FULL" {
		stateVal = 1.0
	}
	ch <- gauge(c.state, stateVal, imp.Target, mdc, imp.CurrentConnection)

	ops, err := proc.ParseMDStats(filepath.Join(dir, "md_stats"))
	if err != nil {
		logWarn(fmt.Sprintf("md_stats %s", dir), err)
	} else {
		for op, count := range ops {
			ch <- counter(c.ops, float64(count), imp.Target, mdc, op)
		}
	}

	rpc, err := proc.ParseRPCStatsHeader(filepath.Join(dir, "rpc_stats"))
	if err != nil {
		logWarn(fmt.Sprintf("rpc_stats %s", dir), err)
	} else {
		ch <- gauge(c.modifyRPCs, float64(rpc.WriteRPCsInFlight), imp.Target, mdc)
	}

	return nil
}
