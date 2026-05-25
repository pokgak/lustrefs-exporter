package proc

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ImportInfo holds state and connection info from an OSC/MDC import file.
type ImportInfo struct {
	Target            string
	State             string
	CurrentConnection string
}

// ParseImport parses a Lustre osc/mdc import file.
func ParseImport(path string) (*ImportInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := &ImportInfo{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if k, v, ok := strings.Cut(line, ":"); ok {
			key := strings.TrimSpace(k)
			val := strings.Trim(strings.TrimSpace(v), `"`)
			switch key {
			case "target":
				info.Target = val
			case "state":
				info.State = val
			case "current_connection":
				info.CurrentConnection = val
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}
	return info, nil
}
