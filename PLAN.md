# lustrefs-exporter

Prometheus exporter for Lustre client-side metrics. Targets Lustre 2.16.x on Ubuntu 22.04.
Reads directly from `/proc/fs/lustre/` and `lnetctl` — no `lctl` dependency for procfs paths.

## Scope

Client-side only. No `llite` (not present on this Lustre version/config), so no per-mount
read/write throughput. What we do expose:

| Source | What it tells you |
|--------|-------------------|
| `osc/*/import` | OST connection health (FULL vs CONNECTING), target server IP |
| `osc/*/osc_stats` | Lockless (DIO) read/write bytes |
| `osc/*/rpc_stats` | RPCs in flight, pending pages, pages-per-RPC histogram |
| `osc/*/cur_grant_bytes` | Write grant held from OST |
| `osc/*/osc_cached_mb` | Client-side cache: used_mb, busy_cnt, reclaim count |
| `osc/*/unstable_stats` | Dirty write backlog: unstable pages and MB |
| `mdc/*/md_stats` | Metadata op counters: close, create, getattr, intent_lock, setattr, unlink, ... |
| `mdc/*/rpc_stats` | MDC modify RPCs in flight |
| `lnetctl stats show` | Global LNet: send/recv counts, errors, resends, timeouts |
| `lnetctl net show` | Per-NI status (up/down), NID |

## Metric Names

All metrics prefixed `lustre_`.

### OSC (one series per OST, labels: `target`, `ost`)

```
lustre_osc_state{target="pfs0-OST0000_UUID", ost="pfs0-OST0000", server="10.200.8.17@tcp"} 1
# 1 = FULL, 0 = anything else (CONNECTING, RECOVERING, ...)

lustre_osc_lockless_read_bytes_total{target, ost}
lustre_osc_lockless_write_bytes_total{target, ost}

lustre_osc_grant_bytes{target, ost}
lustre_osc_cached_mb{target, ost, state}        # state: used, busy, unevict
lustre_osc_cache_reclaim_total{target, ost}
lustre_osc_unstable_pages{target, ost}
lustre_osc_unstable_mb{target, ost}

lustre_osc_read_rpcs_in_flight{target, ost}
lustre_osc_write_rpcs_in_flight{target, ost}
lustre_osc_dio_rpcs_in_flight{target, ost}
lustre_osc_pending_read_pages{target, ost}
lustre_osc_pending_write_pages{target, ost}
# v2: lustre_osc_read/write_rpc_pages_bucket (pages-per-RPC histogram from rpc_stats table)
```

### MDC (one series per MDT, labels: `target`, `mdc`)

```
lustre_mdc_state{target="pfs0-MDT0000_UUID", mdc="pfs0-MDT0000", server="..."} 1

lustre_mdc_ops_total{target, mdc, operation}    # close, create, getattr, intent_lock, ...
lustre_mdc_modify_rpcs_in_flight{target, mdc}
```

### LNet (node-wide)

```
lustre_lnet_send_total
lustre_lnet_recv_total
lustre_lnet_errors_total
lustre_lnet_resend_total
lustre_lnet_response_timeout_total
lustre_lnet_network_timeout_total
lustre_lnet_ni_status{nid, net_type, interface}  # 1 = up, 0 = down
```

## File Layout

```
lustrefs-exporter/
├── main.go                  # flag parsing, HTTP server, signal handling, startup check
├── collector/
│   ├── collector.go         # Collector interface, registry wiring
│   ├── osc.go               # OSC scraper (import, osc_stats, rpc_stats header, cache, grant)
│   ├── mdc.go               # MDC scraper (md_stats, rpc_stats header, import)
│   └── lnet.go              # lnetctl stats show + net show (concurrent, NI cached 30s)
├── proc/
│   ├── import.go            # parse osc/mdc import files → state + connection
│   ├── kvfile.go            # parseKeyValueFile(path) map[string]int64 — shared by osc_stats,
│   │                        #   cur_grant_bytes, osc_cached_mb, unstable_stats
│   ├── rpc_stats.go         # parse rpc_stats header only (RPCs in flight, pending pages)
│   │                        #   histogram deferred to v2
│   ├── md_stats.go          # parse mdc md_stats (op name + sample count)
│   └── lnetctl.go           # run lnetctl, parse YAML output
├── go.mod
└── go.sum
```

## Key Design Decisions

**procfs paths only, no `lctl`** — reading files directly avoids spawning a subprocess per
scrape and removes the `lctl` dependency. Exception: `lnetctl stats show` and `lnetctl net show`
have no procfs equivalent (LNet stats are not in `/proc/fs/lustre/`).

**Glob discovery** — OSC/MDC targets are discovered at scrape time by globbing
`/proc/fs/lustre/osc/*/` and `/proc/fs/lustre/mdc/*/`. Target name is extracted from the
directory name (e.g. `pfs0-OST0000-osc-ff331f65467c7800` → ost `pfs0-OST0000`).

**Single key-value parser** — `osc_stats`, `cur_grant_bytes`, `osc_cached_mb`, and
`unstable_stats` all use the same `key: value` or `key\tvalue` format. One shared
`parseKeyValueFile(path) map[string]int64` covers all four; callers pick the keys they want.

**lnetctl: two calls, one cached** — `lnetctl` only outputs YAML (no JSON/XML flag). Use:
- `lnetctl stats show` — global counters (errors, resend, timeouts). Fresh per scrape.
- `lnetctl export` — per-NI stats (send/recv/drop per NI) + NI status in one call.
  Cache with 30s TTL since NI topology rarely changes.
Run them concurrently. No simpler format exists.

**rpc_stats histogram deferred** — the pages-per-RPC table is the most complex parse and
least actionable for v1. Collect only the header values (RPCs in flight, pending pages).
Add histogram in v2 if needed. Note: when implemented, the table's `cum %` column means
buckets must be fed as cumulative counts (count_up_to_le), computed as `(cum_pct/100) * total_rpcs`.

**Fail open per-file** — if a single OST's file fails to read (OST evicted, procfs race),
log a warning and continue. Don't fail the whole scrape.

**Graceful no-Lustre startup** — if `--lustre-path` doesn't exist at startup, log a warning
and serve an empty `/metrics`. This keeps the process alive if the Lustre mount is lost
after initial deploy (rather than crash-looping).

**No caching for procfs** — procfs reads are cheap (~µs). Scrape fresh on every request.

## Parse Formats (reference)

### import file (state + connection)
```
    target: pfs0-OST0000_UUID
    state: FULL
    connection:
       current_connection: "10.200.8.17@tcp"
```
Parse with: scan for `state:`, `target:`, `current_connection:` lines.

### osc_stats
```
snapshot_time:            1779683025.377780328 secs.nsecs
lockless_write_bytes      0
lockless_read_bytes       0
```
Parse with: `strings.Cut(line, "\t")` or whitespace split, skip `*_time` lines.

### rpc_stats (header section)
```
read RPCs in flight:  0
write RPCs in flight: 0
DIO RPCs in flight: 0
pending write pages:  0
pending read pages:   0

            read            write
pages per rpc         rpcs   % cum % |       rpcs   % cum %
1:           15164  18  18   |         57   1   1
...
```
Parse header with line prefix matching. Parse histogram table: split on `|`, column 0 = page
count (strip `:`), column 1 of each side = rpc count.

### md_stats
```
close                     1638616 samples [reqs]
intent_lock               24227828 samples [reqs]
```
Parse: `fields[0]` = operation, `fields[1]` = count.

### lnetctl stats show (YAML)
```yaml
statistics:
    msgs_alloc: 0
    send_count: 72420754
    errors: 0
    resend_count: 0
    response_timeout_count: 2
    network_timeout_count: 16
    recv_count: 63612337
```
Parse with `gopkg.in/yaml.v3` into a flat struct.

### lnetctl export (YAML) — single call for NI status + per-NI stats
```yaml
net:
-   net type: tcp1345
    local NI(s):
    -   nid: 10.20.0.8@tcp1345
        status: up
        interfaces:
            0: bond0
        statistics:
            send_count: 72420754
            recv_count: 63612337
            drop_count: 0
        health stats:
            fatal_error: 0
            health value: 0
            timeouts: 0
            error: 0
```
Parse into typed structs to extract nid, status, net type, interface name, and per-NI counters.
Note: `lnetctl export` also includes peer entries — filter to `local NI(s)` blocks only.

## Build & Run

```bash
go build -o lustrefs-exporter .
./lustrefs-exporter --port 32221 --lustre-path /proc/fs/lustre
```

Flags:
- `--port` (default `32221`) — listen port (matches whamcloud default for drop-in compatibility)
- `--lustre-path` (default `/proc/fs/lustre`) — override for testing with fixture dirs
- `--lnetctl` (default `lnetctl`) — path to lnetctl binary

## Testing Strategy

**Fixture-based** — copy real `/proc/fs/lustre/osc/<ost>/` and `mdc/` directories to
`testdata/` and run parsers against them. No Lustre installation needed in CI.

Capture fixtures from the cluster:
```bash
ansible ... -m fetch -a "src=/proc/fs/lustre/osc/pfs0-OST0000-osc-ff331f65467c7800/osc_stats dest=testdata/ flat=no"
```

Use `--lustre-path testdata/proc/fs/lustre` flag to point the exporter at fixture data in tests.
