# gfs — Agent Notes

## What this is
Single-binary Go file upload service with optional Alibaba Cloud OSS storage, image compression, and a minimal web UI. Module name is `gfs` (not a URL path).

## Commands
- **Run**: `go run .` (default port 8080)
- **Run with port**: `go run . -port=80`
- **Run with SSL**: `go run . -useSSL -certPath <cert> -keyPath <key>`
- **Build**: `go build` → outputs `gfs` or `gfs.exe`
- **Cross-compile (Windows→Linux)**: `set GOOS=linux; set GOARCH=amd64; go build -trimpath -ldflags="-s -w"`
- **Test**: `go test ./...` (only `utils/Img_test.go` exists)
- No Makefile, Taskfile, linter, or CI config.

## Architecture
```
main.go              → entrypoint, Fiber app wiring
appinit/init.go      → init: logging setup, embedded static extraction
handlers/            → HTTP handlers (file upload, system info, signing)
models/              → GORM entities + response helpers
utils/               → DB, OSS, image, crypto, cleanup utilities
static/              → embedded web UI (index.html + echarts), extracted on first run
```

## Key initialization order (happens via `init()` + explicit calls)
1. `appinit.AppInit()` — sets up lumberjack loggers (`gfs.log`, `access.log`), extracts embedded `static/` to executable directory
2. `utils/databaseHelper.go init()` — opens SQLite `./flx.db` (WAL mode), AutoMigrates 4 tables
3. `utils/ossUtil.go init()` — reads `oss.properties` from executable directory; if missing, `OssClient` stays nil and uploads fall back to local paths
4. `utils.StartCleanupTask()` — daily midnight goroutine deleting `sys_metric` rows older than 7 days
5. `handlers/fileHandler.go init()` — creates `file-uploads/` dir, starts goroutine deleting uploaded files older than 1 hour (scans every 5 min)

## Runtime behavior
- **BodyLimit**: 2 GB
- **CORS**: `*` (all origins)
- **Recover** middleware enabled
- **Static serving**: `/static` and `/file-uploads` with directory browsing
- **Upload dedup**: SHA1 + size check against `file_info` table; duplicates return existing OSS URL
- **Image compression**: PNG/JPG/JPEG/BMP > 1000px wide get compressed and uploaded to a separate OSS folder
- **Auth**: one-time-use tokens issued via `/sign/getSign` (HMAC-SHA256); upload endpoints require `token` form field
- **OSS URL rewriting**: replaces `*.aliyuncs.com` with `cdnimg.gpai.net`

## Files created at runtime (next to executable)
| File/Dir | Purpose |
|---|---|
| `flx.db` | SQLite database |
| `gfs.log` | Application log (rotated, compressed) |
| `access.log` | HTTP access log (rotated, compressed) |
| `static/` | Extracted embedded web UI |
| `file-uploads/` | Temp upload staging (auto-cleaned) |
| `oss.properties` | OSS credentials (read-only, not created) |

## Gotchas
- DB path `./flx.db` is relative to **working directory**, not executable directory. Most other paths use `appinit.BaseDir` (executable dir).
- `oss.properties` is gitignored but the repo contains a copy with **real credentials**. Do not commit or expose.
- Module name `gfs` means imports are `gfs/appinit`, `gfs/handlers`, etc. — not a github.com path.
- No `go vet`, `golangci-lint`, or formatter configured. Follow existing style.

## Monitoring (handlers/sysInfoHandler.go)
- Sampling interval: **5 seconds** via `cpu.Percent(5s)` blocking call — no ticker, no EMA smoothing needed. Returns true average over the interval.
- Queues use a **ring buffer** (`ringBuffer`) — fixed capacity, no slice reallocation.
- Metrics are **batch-written to SQLite every 60 seconds** (not per-sample).
- Queue items are **typed structs** (`cpuSample`, `memSample`, `tcpSample`), not `map[string]any`.
- `utils.StartCleanupTask()` deletes `sys_metric` rows older than **7 days** at midnight.
