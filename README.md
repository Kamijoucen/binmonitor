# binmonitor

跨平台文件监控控制台工具，递归监控指定目录下所有文件的变化，实时输出包含事件类型、文件路径、大小变化的人类可读记录。

本项目遵循 [Atomic Architecture](https://github.com/Kamijoucen/atomic-architecture-md/blob/master/architecture.md) 架构规范组织代码。

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
# 在当前位置初始化默认配置文件 binmonitor.json
./binmonitor init

# 默认读取当前目录 binmonitor.json；配置不存在时监控当前目录
./binmonitor

# 监控指定目录
./binmonitor /path/to/watch

# 使用指定配置文件
./binmonitor -config /path/to/binmonitor.json
```

## 配置文件

配置文件使用 JSON 格式，默认文件名为 `binmonitor.json`。

```json
{
	"root": ".",
	"ignore": [
		"logs",
		"tmp/cache.db"
	]
}
```

字段说明：

- `root`：监控目录，可以是相对路径或绝对路径。
- `ignore`：忽略路径列表，每一项都相对于 `root`。忽略项使用精确路径匹配；如果忽略项是目录，会忽略该目录及其所有子路径。

例如上面的配置会监控当前目录，同时忽略 `logs` 目录下的所有事件，以及 `tmp/cache.db` 文件。

## 输出示例

```
2026-05-28 12:21:55 CREATE /tmp/test/file1.txt 0B → 6B (6B)
2026-05-28 12:21:55 WRITE  /tmp/test/file1.txt 6B → 12B (6B)
2026-05-28 12:21:56 RENAME /tmp/test/file1.txt 12B → 0B (-12B)
2026-05-28 12:21:56 REMOVE /tmp/test/file2.txt 12B → 0B (-12B)
```
