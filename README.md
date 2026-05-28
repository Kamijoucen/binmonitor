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
	],
	"events": [
		"create",
		"write",
		"remove",
		"rename"
	]
}
```

字段说明：

- `root`：监控目录，可以是相对路径或绝对路径。
- `ignore`：忽略路径列表，每一项都相对于 `root`。忽略项使用精确路径匹配；如果忽略项是目录，会忽略该目录及其所有子路径。
- `events`：需要输出的事件类型。支持 `create`、`write`、`remove`、`rename`、`read`，也支持 `modify` 作为 `write` 的别名、`delete` 作为 `remove` 和 `rename` 的组合别名。

例如上面的配置会监控当前目录，同时忽略 `logs` 目录下的所有事件，以及 `tmp/cache.db` 文件。

读取事件基于 Linux inotify 的 `IN_ACCESS` 实现，仅在 Linux/Android 环境可用。其他平台配置 `read` 时会在启动阶段提示不支持。

如果只想监控创建、修改、删除和读取，可以这样配置：

```json
{
	"root": ".",
	"ignore": [],
	"events": [
		"create",
		"modify",
		"delete",
		"read"
	]
}
```

## 输出示例

```
2026-05-28 12:21:55 CREATE /tmp/test/file1.txt 0B → 6B (6B)
2026-05-28 12:21:55 WRITE  /tmp/test/file1.txt 6B → 12B (6B)
2026-05-28 12:21:56 RENAME /tmp/test/file1.txt 12B → 0B (-12B)
2026-05-28 12:21:56 REMOVE /tmp/test/file2.txt 12B → 0B (-12B)
2026-05-28 12:21:57 READ /tmp/test/file3.txt 8B → 8B (0B)
```
