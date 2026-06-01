package types

const (
	// ModeDirectory 表示递归目录监控模式。
	ModeDirectory = "directory"
	// ModeProcess 表示进程文件操作监控模式。
	ModeProcess = "process"
	// DefaultProcessPollIntervalMs 是进程 FD 轮询的默认间隔。
	DefaultProcessPollIntervalMs = 200
)

// Config 表示 binmonitor 的启动配置。
type Config struct {
	Mode                  string          `json:"mode,omitempty"`
	Root                  string          `json:"root"`
	Ignore                []string        `json:"ignore"`
	Events                []string        `json:"events"`
	Log                   bool            `json:"log"`
	DedupLog              bool            `json:"dedupLog"`
	ProcessPollIntervalMs int             `json:"processPollIntervalMs,omitempty"`
	Process               *ProcessConfig  `json:"process,omitempty"`
	Processes             []ProcessConfig `json:"processes,omitempty"`
	EventsConfigured      bool            `json:"-"`
}

// ProcessConfig 表示单个目标进程的监控配置。
type ProcessConfig struct {
	PID            int    `json:"pid"`
	Name           string `json:"name,omitempty"`
	PollIntervalMs int    `json:"pollIntervalMs,omitempty"`
}
