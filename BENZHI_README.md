# bloom-filter — Go 布隆过滤成员判定 HTTP 服务，支持 CRC 快照与崩溃恢复

本成员判定 HTTP 服务维护布隆过滤器：给定元素做增查，快照带 CRC 校验；追加日志可从检查点崩溃恢复；非法输入须报错，不得静默丢位。

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
