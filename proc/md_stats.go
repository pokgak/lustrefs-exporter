package proc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseMDStats parses a Lustre mdc md_stats file.
// Returns a map of operation name → sample count.
func ParseMDStats(path string) (map[string]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasSuffix(line, "secs.nsecs") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		op := fields[0]
		count, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}
		result[op] = count
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}
	return result, nil
}
