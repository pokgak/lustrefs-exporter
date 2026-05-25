package collector

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pokgak/lustrefs-exporter/proc"
)

type oscCollector struct {
	lustrePath string

	state           *prometheus.Desc
	locklessReadBytes  *prometheus.Desc
	locklessWriteBytes *prometheus.Desc
	grantBytes      *prometheus.Desc
	cachedMB        *prometheus.Desc
	cacheReclaim    *prometheus.Desc
	unstablePages   *prometheus.Desc
	unstableMB      *prometheus.Desc
	readRPCs        *prometheus.Desc
	writeRPCs       *prometheus.Desc
	dioRPCs         *prometheus.Desc
	pendingReadPages *prometheus.Desc
	pendingWritePages *prometheus.Desc
}

var oscLabels = []string{"target", "ost"}

func newOSCCollector(lustrePath string) *oscCollector {
	d := func(name, help string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc("lustre_"+name, help, labels, nil)
	}
	return &oscCollector{
		lustrePath:         lustrePath,
		state:              d("osc_state", "OSC connection state (1=FULL, 0=other)", append(oscLabels, "server")...),
		locklessReadBytes:  d("osc_lockless_read_bytes_total", "Lockless (DIO) read bytes", oscLabels...),
		locklessWriteBytes: d("osc_lockless_write_bytes_total", "Lockless (DIO) write bytes", oscLabels...),
		grantBytes:         d("osc_grant_bytes", "Write grant bytes held from OST", oscLabels...),
		cachedMB:           d("osc_cached_mb", "Client-side cache MB", append(oscLabels, "state")...),
		cacheReclaim:       d("osc_cache_reclaim_total", "Cache reclaim count", oscLabels...),
		unstablePages:      d("osc_unstable_pages", "Unstable dirty write backlog pages", oscLabels...),
		unstableMB:         d("osc_unstable_mb", "Unstable dirty write backlog MB", oscLabels...),
		readRPCs:           d("osc_read_rpcs_in_flight", "Read RPCs in flight", oscLabels...),
		writeRPCs:          d("osc_write_rpcs_in_flight", "Write RPCs in flight", oscLabels...),
		dioRPCs:            d("osc_dio_rpcs_in_flight", "DIO RPCs in flight", oscLabels...),
		pendingReadPages:   d("osc_pending_read_pages", "Pending read pages", oscLabels...),
		pendingWritePages:  d("osc_pending_write_pages", "Pending write pages", oscLabels...),
	}
}

func (c *oscCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.state
	ch <- c.locklessReadBytes
	ch <- c.locklessWriteBytes
	ch <- c.grantBytes
	ch <- c.cachedMB
	ch <- c.cacheReclaim
	ch <- c.unstablePages
	ch <- c.unstableMB
	ch <- c.readRPCs
	ch <- c.writeRPCs
	ch <- c.dioRPCs
	ch <- c.pendingReadPages
	ch <- c.pendingWritePages
}

func (c *oscCollector) collect(ch chan<- prometheus.Metric) {
	dirs, err := filepath.Glob(filepath.Join(c.lustrePath, "osc", "*"))
	if err != nil || len(dirs) == 0 {
		return
	}

	for _, dir := range dirs {
		name := filepath.Base(dir)
		ost := oscDirToOST(name)
		if err := c.collectOST(ch, dir, ost); err != nil {
			logWarn(fmt.Sprintf("osc %s", name), err)
		}
	}
}

// oscDirToOST extracts the OST name from a directory like pfs0-OST0000-osc-ff331f65467c7800.
func oscDirToOST(dir string) string {
	parts := strings.Split(dir, "-osc-")
	if len(parts) >= 1 {
		return parts[0]
	}
	return dir
}

func gauge(desc *prometheus.Desc, v float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, v, labels...)
}

func counter(desc *prometheus.Desc, v float64, labels ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc, prometheus.CounterValue, v, labels...)
}

func (c *oscCollector) collectOST(ch chan<- prometheus.Metric, dir, ost string) error {
	imp, err := proc.ParseImport(filepath.Join(dir, "import"))
	if err != nil {
		return fmt.Errorf("import: %w", err)
	}

	stateVal := 0.0
	if imp.State == "FULL" {
		stateVal = 1.0
	}
	ch <- gauge(c.state, stateVal, imp.Target, ost, imp.CurrentConnection)

	stats, err := proc.ParseKeyValueFile(filepath.Join(dir, "osc_stats"))
	if err != nil {
		logWarn(fmt.Sprintf("osc_stats %s", dir), err)
	} else {
		ch <- counter(c.locklessReadBytes, float64(stats["lockless_read_bytes"]), imp.Target, ost)
		ch <- counter(c.locklessWriteBytes, float64(stats["lockless_write_bytes"]), imp.Target, ost)
	}

	grantBytes, err := proc.ReadSingleInt64(filepath.Join(dir, "cur_grant_bytes"))
	if err != nil {
		logWarn(fmt.Sprintf("cur_grant_bytes %s", dir), err)
	} else {
		ch <- gauge(c.grantBytes, float64(grantBytes), imp.Target, ost)
	}

	cached, err := proc.ParseKeyValueFile(filepath.Join(dir, "osc_cached_mb"))
	if err != nil {
		logWarn(fmt.Sprintf("osc_cached_mb %s", dir), err)
	} else {
		ch <- gauge(c.cachedMB, float64(cached["used_mb"]), imp.Target, ost, "used")
		ch <- gauge(c.cachedMB, float64(cached["busy_cnt"]), imp.Target, ost, "busy")
		ch <- gauge(c.cachedMB, float64(cached["unevict"]), imp.Target, ost, "unevict")
		ch <- counter(c.cacheReclaim, float64(cached["reclaim"]), imp.Target, ost)
	}

	unstable, err := proc.ParseKeyValueFile(filepath.Join(dir, "unstable_stats"))
	if err != nil {
		logWarn(fmt.Sprintf("unstable_stats %s", dir), err)
	} else {
		ch <- gauge(c.unstablePages, float64(unstable["unstable_pages"]), imp.Target, ost)
		ch <- gauge(c.unstableMB, float64(unstable["unstable_mb"]), imp.Target, ost)
	}

	rpc, err := proc.ParseRPCStatsHeader(filepath.Join(dir, "rpc_stats"))
	if err != nil {
		logWarn(fmt.Sprintf("rpc_stats %s", dir), err)
	} else {
		ch <- gauge(c.readRPCs, float64(rpc.ReadRPCsInFlight), imp.Target, ost)
		ch <- gauge(c.writeRPCs, float64(rpc.WriteRPCsInFlight), imp.Target, ost)
		ch <- gauge(c.dioRPCs, float64(rpc.DIORPCsInFlight), imp.Target, ost)
		ch <- gauge(c.pendingReadPages, float64(rpc.PendingReadPages), imp.Target, ost)
		ch <- gauge(c.pendingWritePages, float64(rpc.PendingWritePages), imp.Target, ost)
	}

	return nil
}
