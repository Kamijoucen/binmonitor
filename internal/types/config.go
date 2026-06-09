package types

const (
	// ModeDirectory 表示递归目录监控模式。
	ModeDirectory = "directory"
	// ModeProcess 表示进程文件操作监控模式。
	ModeProcess = "process"
	// ModeNetwork 表示进程网络连接监控模式。
	ModeNetwork = "network"
	// DefaultProcessPollIntervalMs 是进程 FD 轮询的默认间隔。
	DefaultProcessPollIntervalMs = 200
	// DefaultNetworkPollIntervalMs 是网络连接轮询的默认间隔。
	DefaultNetworkPollIntervalMs = 500
)

// DefaultSocks5Ports 是 SOCKS5 代理常用端口列表。
var DefaultSocks5Ports = []int{1080, 10800, 9050}

// Config 表示 binmonitor 的启动配置。
type Config struct {
	Mode                  string            `json:"mode,omitempty"`
	Root                  string            `json:"root"`
	Ignore                []string          `json:"ignore"`
	Events                []string          `json:"events"`
	Log                   bool              `json:"log"`
	DedupLog              bool              `json:"dedupLog"`
	ProcessPollIntervalMs int               `json:"processPollIntervalMs,omitempty"`
	Process               *ProcessConfig    `json:"process,omitempty"`
	Processes             []ProcessConfig   `json:"processes,omitempty"`
	NetMonitor            *NetMonitorConfig `json:"netMonitor,omitempty"`
	EventsConfigured      bool              `json:"-"`
}

// ProcessConfig 表示单个目标进程的监控配置。
type ProcessConfig struct {
	PID            int    `json:"pid"`
	Name           string `json:"name,omitempty"`
	PollIntervalMs int    `json:"pollIntervalMs,omitempty"`
}

// NetMonitorConfig 表示网络连接监控的配置。
type NetMonitorConfig struct {
	PID            int   `json:"pid"`            // 目标进程 PID
	PollIntervalMs int   `json:"pollIntervalMs"` // 轮询间隔（毫秒），默认 500
	DNSTrace       bool  `json:"dnsTrace"`       // 是否检测 DNS 查询（dport=53 的 UDP 连接）
	Socks5Ports    []int `json:"socks5Ports"`    // SOCKS5 代理端口列表
}
