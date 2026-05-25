# lustrefs-exporter

Prometheus exporter for Lustre client-side metrics. Reads directly from `/proc/fs/lustre/` and `lnetctl` — no `lctl` dependency.

## Supported versions

| Lustre | Ubuntu |
|--------|--------|
| 2.16.x | 22.04 (Jammy), 24.04 (Noble) |

Client-side only. No `llite` metrics (not present in this Lustre version/config), so no per-mount read/write throughput.

## Metrics

All metrics are prefixed `lustre_`.

### OSC — one series per OST

Labels: `target` (UUID), `ost` (short name), `server` (IP@net, state metric only)

| Metric | Type | Description |
|--------|------|-------------|
| `lustre_osc_state` | Gauge | Connection state: 1 = FULL, 0 = other (CONNECTING, RECOVERING, …) |
| `lustre_osc_lockless_read_bytes_total` | Counter | Lockless (DIO) read bytes |
| `lustre_osc_lockless_write_bytes_total` | Counter | Lockless (DIO) write bytes |
| `lustre_osc_grant_bytes` | Gauge | Write grant bytes held from OST |
| `lustre_osc_cached_mb` | Gauge | Client-side cache MB — label `state`: `used`, `busy`, `unevict` |
| `lustre_osc_cache_reclaim_total` | Counter | Cache reclaim count |
| `lustre_osc_unstable_pages` | Gauge | Dirty write backlog: unstable pages |
| `lustre_osc_unstable_mb` | Gauge | Dirty write backlog: unstable MB |
| `lustre_osc_read_rpcs_in_flight` | Gauge | Read RPCs currently in flight |
| `lustre_osc_write_rpcs_in_flight` | Gauge | Write RPCs currently in flight |
| `lustre_osc_dio_rpcs_in_flight` | Gauge | DIO RPCs currently in flight |
| `lustre_osc_pending_read_pages` | Gauge | Pending read pages |
| `lustre_osc_pending_write_pages` | Gauge | Pending write pages |

### MDC — one series per MDT

Labels: `target` (UUID), `mdc` (short name), `server` (IP@net, state metric only)

| Metric | Type | Description |
|--------|------|-------------|
| `lustre_mdc_state` | Gauge | Connection state: 1 = FULL, 0 = other |
| `lustre_mdc_ops_total` | Counter | Metadata op count — label `operation`: `close`, `create`, `getattr`, `intent_lock`, `setattr`, `unlink`, … |
| `lustre_mdc_modify_rpcs_in_flight` | Gauge | Modify RPCs currently in flight |

### LNet — node-wide

| Metric | Type | Description |
|--------|------|-------------|
| `lustre_lnet_send_total` | Counter | Total messages sent |
| `lustre_lnet_recv_total` | Counter | Total messages received |
| `lustre_lnet_errors_total` | Counter | Total errors |
| `lustre_lnet_resend_total` | Counter | Total resends |
| `lustre_lnet_response_timeout_total` | Counter | Response timeouts |
| `lustre_lnet_network_timeout_total` | Counter | Network timeouts |
| `lustre_lnet_ni_status` | Gauge | Per-NI status: 1 = up, 0 = down — labels: `nid`, `net_type`, `interface` |

## Installation

### apt (Ubuntu 22.04 / 24.04)

Download the latest `.deb` from the [releases page](https://github.com/pokgak/lustrefs-exporter/releases) and install:

```bash
wget https://github.com/pokgak/lustrefs-exporter/releases/latest/download/lustrefs-exporter_<version>_linux_amd64.deb
sudo dpkg -i lustrefs-exporter_<version>_linux_amd64.deb
sudo systemctl start lustrefs-exporter
```

The service is enabled automatically on install. Default listen port is `32221`.

### Configuration

Edit `/etc/default/lustrefs-exporter` to override flags:

```bash
LUSTREFS_EXPORTER_OPTS="--port 32221 --lustre-path /proc/fs/lustre"
```

Then restart: `sudo systemctl restart lustrefs-exporter`

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `32221` | Listen port |
| `--lustre-path` | `/proc/fs/lustre` | Path to Lustre procfs (override for testing) |
| `--lnetctl` | `lnetctl` | Path to `lnetctl` binary |

Metrics are served at `http://<host>:32221/metrics`.

## Building from source

Requires Go 1.22+ and [nfpm](https://nfpm.goreleaser.com/) for packaging.

```bash
# Run tests
make test

# Build Linux amd64 binary → dist/lustrefs-exporter
make build

# Local snapshot build (.deb + binary, no publish)
make snapshot
```

## Releasing

Push a version tag to trigger a GitHub Actions build that publishes a release with the `.deb`, `.tar.gz`, and `checksums.txt`:

```bash
git tag v0.1.0
git push --tags
```

## Architecture

- **procfs-only for OSC/MDC** — reads `/proc/fs/lustre/osc/*/` and `mdc/*/` directly; no subprocess per scrape
- **lnetctl for LNet** — `lnetctl stats show` (fresh per scrape) and `lnetctl export` (NI status, cached 30 s) run concurrently
- **Fail open per-file** — if one OST's files fail (evicted, procfs race), a warning is logged and the rest of the scrape continues
- **Graceful no-Lustre startup** — if `--lustre-path` doesn't exist at startup, the process serves empty metrics rather than crash-looping
- **Vendored dependencies** — `vendor/` is committed; builds are fully offline and reproducible
