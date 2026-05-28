package types

// Config 表示 binmonitor 的启动配置。
type Config struct {
	Root   string   `json:"root"`
	Ignore []string `json:"ignore"`
	Events []string `json:"events"`
	Log    bool     `json:"log"`
}
