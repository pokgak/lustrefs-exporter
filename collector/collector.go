package collector

import (
	"log"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

// LustreCollector aggregates OSC, MDC and LNet sub-collectors.
type LustreCollector struct {
	lustre  string
	lnetctl string
	osc     *oscCollector
	mdc     *mdcCollector
	lnet    *lnetCollector
}

// New creates a LustreCollector.
func New(lustrePath, lnetctlBin string) *LustreCollector {
	return &LustreCollector{
		lustre:  lustrePath,
		lnetctl: lnetctlBin,
		osc:     newOSCCollector(lustrePath),
		mdc:     newMDCCollector(lustrePath),
		lnet:    newLNetCollector(lnetctlBin),
	}
}

// Describe implements prometheus.Collector.
func (c *LustreCollector) Describe(ch chan<- *prometheus.Desc) {
	c.osc.describe(ch)
	c.mdc.describe(ch)
	c.lnet.describe(ch)
}

// Collect implements prometheus.Collector.
func (c *LustreCollector) Collect(ch chan<- prometheus.Metric) {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); c.osc.collect(ch) }()
	go func() { defer wg.Done(); c.mdc.collect(ch) }()
	go func() { defer wg.Done(); c.lnet.collect(ch) }()
	wg.Wait()
}

func logWarn(msg string, err error) {
	log.Printf("WARN %s: %v", msg, err)
}
