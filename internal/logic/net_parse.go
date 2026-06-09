// Package logic 中的 /proc/net/* 文件解析与网络事件分类函数。
// 遵循 Atomic Architecture V2：纯无状态函数，零架构依赖。
package logic

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"binmonitor/internal/types"
)

// ReadProcNetFiles 读取进程的网络连接信息，返回 inode → NetConnInfo 的映射。
func ReadProcNetFiles(pid int) (map[uint64]types.NetConnInfo, error) {
	inodes, err := readProcessSocketInodes(pid)
	if err != nil {
		return nil, fmt.Errorf("read process %d socket inodes: %w", pid, err)
	}
	if len(inodes) == 0 {
		return nil, nil
	}

	conns := make(map[uint64]types.NetConnInfo, len(inodes))

	netFiles := []struct {
		path     string
		protocol string
	}{
		{"/proc/net/tcp", "tcp"},
		{"/proc/net/tcp6", "tcp6"},
		{"/proc/net/udp", "udp"},
		{"/proc/net/udp6", "udp6"},
	}

	for _, nf := range netFiles {
		entries, err := parseProcNetFile(nf.path, nf.protocol)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if _, ok := inodes[entry.Inode]; ok {
				conns[entry.Inode] = entry
			}
		}
	}

	return conns, nil
}

// readProcessSocketInodes 从 /proc/<pid>/fd/ 读取进程持有的 socket inode 集合。
func readProcessSocketInodes(pid int) (map[uint64]struct{}, error) {
	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil, err
	}

	inodes := make(map[uint64]struct{})
	for _, entry := range entries {
		fdPath := filepath.Join(fdDir, entry.Name())
		target, err := os.Readlink(fdPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(target, "socket:[") {
			continue
		}
		inodeStr := target[len("socket:[") : len(target)-1]
		inode, err := strconv.ParseUint(inodeStr, 10, 64)
		if err != nil {
			continue
		}
		inodes[inode] = struct{}{}
	}
	return inodes, nil
}

// parseProcNetFile 解析 /proc/net/tcp 或 /proc/net/udp 文件。
func parseProcNetFile(path string, protocol string) ([]types.NetConnInfo, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []types.NetConnInfo
	scanner := bufio.NewScanner(f)
	firstLine := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if firstLine {
			firstLine = false
			if strings.HasPrefix(line, "sl") {
				continue
			}
		}
		entry, err := ParseProcNetLine(line, protocol)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}
	return entries, scanner.Err()
}

// ParseProcNetLine 解析 /proc/net 中的单行。
func ParseProcNetLine(line string, protocol string) (types.NetConnInfo, error) {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return types.NetConnInfo{}, fmt.Errorf("too few fields in /proc/net line: %s", line)
	}

	localIP, localPort, err := ParseHexAddress(fields[1])
	if err != nil {
		return types.NetConnInfo{}, err
	}
	remoteIP, remotePort, err := ParseHexAddress(fields[2])
	if err != nil {
		return types.NetConnInfo{}, err
	}

	state := fields[3]
	if name, ok := types.TCPStateNames[state]; ok {
		state = name
	}

	inode, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return types.NetConnInfo{}, fmt.Errorf("parse inode: %w", err)
	}

	return types.NetConnInfo{
		Protocol: protocol,
		SrcIP:    localIP,
		SrcPort:  localPort,
		DstIP:    remoteIP,
		DstPort:  remotePort,
		State:    state,
		Inode:    inode,
	}, nil
}

// ParseHexAddress 解析格式为 "hexIP:hexPort" 的地址字符串。
func ParseHexAddress(hexAddr string) (ip string, port uint16, err error) {
	parts := strings.SplitN(hexAddr, ":", 2)
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid hex address: %s", hexAddr)
	}

	port64, err := strconv.ParseUint(parts[1], 16, 16)
	if err != nil {
		return "", 0, fmt.Errorf("parse hex port: %w", err)
	}

	ip = ParseHexIP(parts[0])
	return ip, uint16(port64), nil
}

// ParseHexIP 将 little-endian 十六进制 IP 字符串转为点分十进制。
func ParseHexIP(hex string) string {
	if len(hex) == 8 {
		n, err := strconv.ParseUint(hex, 16, 32)
		if err != nil {
			return hex
		}
		return fmt.Sprintf("%d.%d.%d.%d", byte(n), byte(n>>8), byte(n>>16), byte(n>>24))
	}

	if len(hex) == 32 {
		parts := make([]string, 8)
		for i := 0; i < 8; i++ {
			seg := hex[i*4 : (i+1)*4]
			n, err := strconv.ParseUint(seg, 16, 16)
			if err != nil {
				return hex
			}
			parts[i] = fmt.Sprintf("%x", (n>>8)|((n&0xff)<<8))
		}
		return strings.Join(parts, ":")
	}

	return hex
}

// BuildPortSet 将端口列表转为 set，用于快速查找。
func BuildPortSet(ports []int) map[uint16]struct{} {
	s := make(map[uint16]struct{}, len(ports))
	for _, p := range ports {
		if p > 0 && p <= 65535 {
			s[uint16(p)] = struct{}{}
		}
	}
	return s
}

// ClassifyNetOp 根据连接信息判断操作类型。
func ClassifyNetOp(info types.NetConnInfo, dnsTrace bool, socks5Ports map[uint16]struct{}) types.NetOp {
	// SOCKS5 代理端口检测
	if _, ok := socks5Ports[info.DstPort]; ok {
		return types.OpSOCKS5Connect
	}
	// DNS 查询检测（UDP 目标端口 53）
	if dnsTrace && info.DstPort == 53 && strings.HasPrefix(info.Protocol, "udp") {
		return types.OpDNSQuery
	}
	// UDP 连接
	if strings.HasPrefix(info.Protocol, "udp") {
		return types.OpUDPConnect
	}
	// TCP 连接
	return types.OpTCPConnect
}
