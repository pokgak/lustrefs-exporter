package proc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// RPCStatsHeader holds the header counters from an rpc_stats file.
type RPCStatsHeader struct {
	ReadRPCsInFlight  int64
	WriteRPCsInFlight int64
	DIORPCsInFlight   int64
	PendingWritePages int64
	PendingReadPages  int64
}

// ParseRPCStatsHeader reads rpc_stats and returns only the header counters.
func ParseRPCStatsHeader(path string) (*RPCStatsHeader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := &RPCStatsHeader{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if v, ok := parseRPCHeaderLine(line, "read RPCs in flight:"); ok {
			h.ReadRPCsInFlight = v
		} else if v, ok := parseRPCHeaderLine(line, "write RPCs in flight:"); ok {
			h.WriteRPCsInFlight = v
		} else if v, ok := parseRPCHeaderLine(line, "DIO RPCs in flight:"); ok {
			h.DIORPCsInFlight = v
		} else if v, ok := parseRPCHeaderLine(line, "pending write pages:"); ok {
			h.PendingWritePages = v
		} else if v, ok := parseRPCHeaderLine(line, "pending read pages:"); ok {
			h.PendingReadPages = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}
	return h, nil
}

func parseRPCHeaderLine(line, prefix string) (int64, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, prefix) {
		return 0, false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return 0, false
	}
	v, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}
