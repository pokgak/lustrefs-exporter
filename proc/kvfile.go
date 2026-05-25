package proc

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseKeyValueFile parses files with "key: value" or "key\tvalue" lines.
// Lines that don't parse cleanly (headers, blank lines, non-numeric values) are skipped.
func ParseKeyValueFile(path string) (map[string]int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	result := make(map[string]int64)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var key, rest string
		if k, v, ok := strings.Cut(line, "\t"); ok {
			key = strings.TrimSpace(k)
			rest = strings.TrimSpace(v)
		} else if k, v, ok := strings.Cut(line, ":"); ok {
			key = strings.TrimSpace(k)
			rest = strings.TrimSpace(v)
		} else {
			// whitespace-separated: "key value ..."
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			key = fields[0]
			rest = strings.Join(fields[1:], " ")
		}

		// Take first whitespace-separated token as the value
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		v, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			continue
		}
		result[key] = v
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}
	return result, nil
}

// ReadSingleInt64 reads a file containing a single integer.
func ReadSingleInt64(path string) (int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	v, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing %s: %w", path, err)
	}
	return v, nil
}
