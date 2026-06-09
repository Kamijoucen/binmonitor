package types

import "fmt"

// FileOp 表示文件系统操作的类型。
type FileOp int

const (
	OpCreate FileOp = iota
	OpWrite
	OpRemove
	OpRename
	OpRead
	OpChmod
	OpOpen
	OpClose
)

// FileEvent 是由 WatcherComponent 产生的领域无关的文件系统事件。
type FileEvent struct {
	Path        string
	Op          FileOp
	PID         int
	FD          int
	ProcessName string
}

// NetOp 表示网络连接操作的类型。
type NetOp int

const (
	OpTCPConnect NetOp = iota
	OpTCPClose
	OpTCPStateChange
	OpUDPConnect
	OpDNSQuery
	OpSOCKS5Connect
)

// NetEvent 表示由 NetWatcherComponent 产生的网络连接事件。
type NetEvent struct {
	PID         int    // 进程 PID
	ProcessName string // 进程名
	Protocol    string // "tcp", "tcp6", "udp", "udp6"
	SrcIP       string // 源 IP
	SrcPort     uint16 // 源端口
	DstIP       string // 目标 IP
	DstPort     uint16 // 目标端口
	Op          NetOp  // 操作类型
	FD          int    // socket fd 号
}

// NetOpString 返回 NetOp 对应的输出标签。
func (op NetOp) String() string {
	switch op {
	case OpTCPConnect:
		return "TCP_CONNECT"
	case OpTCPClose:
		return "TCP_CLOSE"
	case OpTCPStateChange:
		return "TCP_STATE"
	case OpUDPConnect:
		return "UDP_CONNECT"
	case OpDNSQuery:
		return "DNS_QUERY"
	case OpSOCKS5Connect:
		return "SOCKS5_CONNECT"
	default:
		return "UNKNOWN"
	}
}

// ConnectionString 返回格式化的连接描述字符串。
func (e *NetEvent) ConnectionString() string {
	return fmt.Sprintf("%s %s:%d → %s:%d", e.Protocol, e.SrcIP, e.SrcPort, e.DstIP, e.DstPort)
}

// NetConnInfo 表示从 /proc/net/* 解析出的网络连接信息。
type NetConnInfo struct {
	Protocol string // "tcp", "tcp6", "udp", "udp6"
	SrcIP    string
	SrcPort  uint16
	DstIP    string
	DstPort  uint16
	State    string // TCP state, e.g. "ESTABLISHED", "LISTEN"
	Inode    uint64
}

// TCPStateNames 映射 /proc/net/tcp 的状态码到可读名称。
var TCPStateNames = map[string]string{
	"01": "ESTABLISHED",
	"02": "SYN_SENT",
	"03": "SYN_RECV",
	"04": "FIN_WAIT1",
	"05": "FIN_WAIT2",
	"06": "TIME_WAIT",
	"07": "CLOSE",
	"08": "CLOSE_WAIT",
	"09": "LAST_ACK",
	"0A": "LISTEN",
	"0B": "CLOSING",
}
