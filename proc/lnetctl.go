package proc

import (
	"fmt"
	"os/exec"

	"gopkg.in/yaml.v3"
)

// LNetStats holds global LNet counters from `lnetctl stats show`.
type LNetStats struct {
	SendCount             int64 `yaml:"send_count"`
	RecvCount             int64 `yaml:"recv_count"`
	Errors                int64 `yaml:"errors"`
	ResendCount           int64 `yaml:"resend_count"`
	ResponseTimeoutCount  int64 `yaml:"response_timeout_count"`
	NetworkTimeoutCount   int64 `yaml:"network_timeout_count"`
}

type lnetStatsWrapper struct {
	Statistics LNetStats `yaml:"statistics"`
}

// LNetNI holds per-NI status from `lnetctl export`.
type LNetNI struct {
	NID       string
	Status    string // "up" or "down"
	NetType   string
	Interface string
}

type lnetExport struct {
	Net []lnetExportNet `yaml:"net"`
}

type lnetExportNet struct {
	NetType string           `yaml:"net type"`
	LocalNI []lnetExportNI   `yaml:"local NI(s)"`
}

type lnetExportNI struct {
	NID        string            `yaml:"nid"`
	Status     string            `yaml:"status"`
	Interfaces map[string]string `yaml:"interfaces"`
}

// RunLNetCtlStats executes `lnetctl stats show` and parses the output.
func RunLNetCtlStats(lnetctlBin string) (*LNetStats, error) {
	out, err := exec.Command(lnetctlBin, "stats", "show").Output()
	if err != nil {
		return nil, fmt.Errorf("lnetctl stats show: %w", err)
	}
	return ParseLNetStats(out)
}

// ParseLNetStats parses YAML output of `lnetctl stats show`.
func ParseLNetStats(data []byte) (*LNetStats, error) {
	var w lnetStatsWrapper
	if err := yaml.Unmarshal(data, &w); err != nil {
		return nil, fmt.Errorf("parsing lnetctl stats: %w", err)
	}
	return &w.Statistics, nil
}

// RunLNetCtlExport executes `lnetctl export` and parses the NI list.
func RunLNetCtlExport(lnetctlBin string) ([]LNetNI, error) {
	out, err := exec.Command(lnetctlBin, "export").Output()
	if err != nil {
		return nil, fmt.Errorf("lnetctl export: %w", err)
	}
	return ParseLNetExport(out)
}

// ParseLNetExport parses YAML output of `lnetctl export`.
func ParseLNetExport(data []byte) ([]LNetNI, error) {
	var exp lnetExport
	if err := yaml.Unmarshal(data, &exp); err != nil {
		return nil, fmt.Errorf("parsing lnetctl export: %w", err)
	}

	var nis []LNetNI
	for _, net := range exp.Net {
		for _, ni := range net.LocalNI {
			iface := ""
			for _, v := range ni.Interfaces {
				iface = v
				break
			}
			nis = append(nis, LNetNI{
				NID:       ni.NID,
				Status:    ni.Status,
				NetType:   net.NetType,
				Interface: iface,
			})
		}
	}
	return nis, nil
}
