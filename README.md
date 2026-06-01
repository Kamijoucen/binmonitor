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

# 监控单个进程打开/关闭文件，Linux/RK3588/root Android 可用
./binmonitor -pid 1234

# 同时监控多个进程，并设置 FD 轮询间隔
./binmonitor -pid 1234,5678 -poll-ms 200
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
	],
	"log": false,
	"dedupLog": false
}
```

字段说明：

- `root`：监控目录，可以是相对路径或绝对路径。
- `ignore`：忽略路径列表，每一项都相对于 `root`。忽略项使用精确路径匹配；如果忽略项是目录，会忽略该目录及其所有子路径。
- `mode`：监控模式。缺省为 `directory`，表示递归监控目录；设为 `process` 时按 PID 监控进程打开/关闭文件。
- `events`：需要输出的事件类型。目录模式支持 `create`、`write`、`remove`、`rename`、`read`，也支持 `modify` 作为 `write` 的别名、`delete` 作为 `remove` 和 `rename` 的组合别名。进程模式支持 `open`、`close`，也支持 `process_open`、`process_close` 别名。
- `log`：是否开启文件日志输出。设为 `true` 时，监控事件和错误信息会同步追加写入当前工作目录下的 `binmonitor.log` 文件中。日志文件以追加模式打开，程序重启不会覆盖历史日志。启用日志后，`binmonitor.log` 会自动加入忽略列表，避免自监控递归。
- `dedupLog`：是否开启去重统计日志。设为 `true` 时，每个文件的每种事件类型在当前进程生命周期内仅记录一次，并实时覆盖写入当前工作目录下的 `binmonitor.dedup.log` 文件中。统计按事件类型分组，组内按首次发生顺序排列。启用后，`binmonitor.dedup.log` 会自动加入忽略列表。
- `processPollIntervalMs`：进程模式的默认 FD 快照轮询间隔，单位毫秒，缺省为 `200`。
- `processes`：进程模式的目标进程列表。每项包含 `pid`、可选 `name` 和可选 `pollIntervalMs`；单项 `pollIntervalMs` 会覆盖全局 `processPollIntervalMs`。`name` 只用于输出标识，不按进程名查找。

例如上面的配置会监控当前目录，同时忽略 `logs` 目录下的所有事件，以及 `tmp/cache.db` 文件。

读取事件基于 Linux inotify 的 `IN_ACCESS` 实现，仅在 Linux/Android 环境可用。其他平台配置 `read` 时会在启动阶段提示不支持。

进程文件操作监控基于 Linux `/proc/<pid>/fd` 定时快照，仅在 Linux/RK3588 和已获取 root 权限的 Android 环境可用。它发现的是目标进程 FD 对应文件的打开和关闭，不区分实际 `read`/`write` syscall，也不保证捕获极短生命周期的打开/关闭操作。若目标 PID 退出、权限不足，或 Android SELinux 限制读取 `/proc/<pid>/fd`，程序会输出对应错误。

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

如果要通过配置文件同时监控多个进程，可以这样配置：

```json
{
	"mode": "process",
	"events": [
		"open",
		"close"
	],
	"processPollIntervalMs": 200,
	"processes": [
		{
			"pid": 1234,
			"name": "api"
		},
		{
			"pid": 5678,
			"name": "worker",
			"pollIntervalMs": 100
		}
	],
	"log": false,
	"dedupLog": false
}
```

命令行传入 `-pid` 时会进入进程模式，并覆盖配置文件中的 `processes` 列表；配置文件里的 `events`、`log`、`dedupLog` 等其他字段仍会生效。

## 输出示例

```
2026-05-28 12:21:55 CREATE /tmp/test/file1.txt 0B → 6B (6B)
2026-05-28 12:21:55 WRITE  /tmp/test/file1.txt 6B → 12B (6B)
2026-05-28 12:21:56 RENAME /tmp/test/file1.txt 12B → 0B (-12B)
2026-05-28 12:21:56 REMOVE /tmp/test/file2.txt 12B → 0B (-12B)
2026-05-28 12:21:57 READ /tmp/test/file3.txt 8B → 8B (0B)
2026-05-28 12:21:58 OPEN pid=1234 fd=7 name=api /tmp/test/file4.txt 4B
2026-05-28 12:21:59 CLOSE pid=1234 fd=7 name=api /tmp/test/file4.txt 4B
```

## 去重统计日志

当 `dedupLog` 设为 `true` 时，除控制台输出外，程序还会生成 `binmonitor.dedup.log`，按事件类型分组记录每个文件首次触发的事件：

```
[CREATE]
/tmp/test/file1.txt
/tmp/test/file2.txt

[WRITE]
/tmp/test/file1.txt

[READ]
/tmp/test/file3.txt

[OPEN]
pid=1234 name=api /tmp/test/file4.txt

[CLOSE]
pid=1234 name=api /tmp/test/file4.txt

[REMOVE]
/tmp/test/file2.txt
```

同一文件的同一事件类型仅记录一次，新事件追加到对应组末尾，每次有新记录时覆盖写入文件。
