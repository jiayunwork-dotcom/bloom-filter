Go 命令行布隆过滤器：add/check 读写带 CRC 的快照文件，store-add 走追加日志并可从检查点崩溃恢复；bloom-filter serve 把同一套仓挂到 HTTP 供增查。

# bloom-filter

Space-efficient probabilistic set membership service with append-only persistent
storage, CRC32 integrity checks, and crash recovery via checkpoint/replay.

## Build / Run / Test

```bash
go build -o bloom-filter .
./bloom-filter serve -addr :8080 -f data.blm
go test ./...
```

## Evaluation Image

Evaluation-specific files (do not overwrite project Dockerfile/README):

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md` (this file)

Build and verify in container:

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```
