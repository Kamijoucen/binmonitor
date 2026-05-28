# binmonitor

跨平台文件监控控制台工具，递归监控指定目录下所有文件的变化，实时输出包含事件类型、文件路径、大小变化的人类可读记录。


## 编译构建

### 本地编译

```bash
go build -o binmonitor ./cmd/binmonitor
```

### RK3588 交叉编译

RK3588 为 ARM64 架构 Linux 环境，使用交叉编译：

```bash
GOOS=linux GOARCH=arm64 go build -o binmonitor-arm64 ./cmd/binmonitor
```

编译完成后将 `binmonitor-arm64` 传输到 RK3588 设备上即可直接运行。

## 运行方式

```bash
# 默认监控当前目录
./binmonitor

# 监控指定目录
./binmonitor /path/to/watch
```

## 输出示例

```
2026-05-28 12:21:55 CREATE /tmp/test/file1.txt 0B → 6B (+6B)
2026-05-28 12:21:55 WRITE  /tmp/test/file1.txt 6B → 12B (+6B)
2026-05-28 12:21:56 RENAME /tmp/test/file1.txt 12B → 0B (-12B)
2026-05-28 12:21:56 REMOVE /tmp/test/file2.txt 12B → 0B (-12B)
```