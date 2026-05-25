package collector

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/pokgak/lustrefs-exporter/proc"
)

type lnetCollector struct {
	lnetctlBin string

	// cached NI list
	mu         sync.Mutex
	niCache    []proc.LNetNI
	niCacheAt  time.Time
	niCacheTTL time.Duration

	send            *prometheus.Desc
	recv            *prometheus.Desc
	errors          *prometheus.Desc
	resend          *prometheus.Desc
	responseTimeout *prometheus.Desc
	networkTimeout  *prometheus.Desc
	niStatus        *prometheus.Desc
}

func newLNetCollector(lnetctlBin string) *lnetCollector {
	d := func(name, help string, labels ...string) *prometheus.Desc {
		return prometheus.NewDesc("lustre_"+name, help, labels, nil)
	}
	return &lnetCollector{
		lnetctlBin: lnetctlBin,
		niCacheTTL: 30 * time.Second,
		send:            d("lnet_send_total", "LNet total send count"),
		recv:            d("lnet_recv_total", "LNet total recv count"),
		errors:          d("lnet_errors_total", "LNet total errors"),
		resend:          d("lnet_resend_total", "LNet total resend count"),
		responseTimeout: d("lnet_response_timeout_total", "LNet response timeout count"),
		networkTimeout:  d("lnet_network_timeout_total", "LNet network timeout count"),
		niStatus:        d("lnet_ni_status", "LNet NI status (1=up, 0=down)", "nid", "net_type", "interface"),
	}
}

func (c *lnetCollector) describe(ch chan<- *prometheus.Desc) {
	ch <- c.send
	ch <- c.recv
	ch <- c.errors
	ch <- c.resend
	ch <- c.responseTimeout
	ch <- c.networkTimeout
	ch <- c.niStatus
}

func (c *lnetCollector) collect(ch chan<- prometheus.Metric) {
	var statsErr, niErr error
	var stats *proc.LNetStats
	var nis []proc.LNetNI

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		stats, statsErr = proc.RunLNetCtlStats(c.lnetctlBin)
	}()
	go func() {
		defer wg.Done()
		nis, niErr = c.getCachedNIs()
	}()
	wg.Wait()

	if statsErr != nil {
		logWarn("lnetctl stats show", statsErr)
	} else {
		ch <- counter(c.send, float64(stats.SendCount))
		ch <- counter(c.recv, float64(stats.RecvCount))
		ch <- counter(c.errors, float64(stats.Errors))
		ch <- counter(c.resend, float64(stats.ResendCount))
		ch <- counter(c.responseTimeout, float64(stats.ResponseTimeoutCount))
		ch <- counter(c.networkTimeout, float64(stats.NetworkTimeoutCount))
	}

	if niErr != nil {
		logWarn("lnetctl export", niErr)
	} else {
		for _, ni := range nis {
			v := 0.0
			if ni.Status == "up" {
				v = 1.0
			}
			ch <- gauge(c.niStatus, v, ni.NID, ni.NetType, ni.Interface)
		}
	}
}

func (c *lnetCollector) getCachedNIs() ([]proc.LNetNI, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if time.Since(c.niCacheAt) < c.niCacheTTL {
		return c.niCache, nil
	}
	nis, err := proc.RunLNetCtlExport(c.lnetctlBin)
	if err != nil {
		return nil, err
	}
	c.niCache = nis
	c.niCacheAt = time.Now()
	return nis, nil
}
