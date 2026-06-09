package logic

import (
	"testing"

	"binmonitor/internal/types"
)

func TestParseHexIP(t *testing.T) {
	tests := []struct {
		hex  string
		want string
	}{
		{"0100007F", "127.0.0.1"},
		{"00000000", "0.0.0.0"},
		{"5C28D893", "147.216.40.92"},
		{"FFFFFFFF", "255.255.255.255"},
		{"COFFEE", "COFFEE"}, // 无效长度，原样返回
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got := ParseHexIP(tt.hex)
			if got != tt.want {
				t.Errorf("ParseHexIP(%q) = %q, want %q", tt.hex, got, tt.want)
			}
		})
	}
}

func TestParseHexAddress(t *testing.T) {
	tests := []struct {
		hexAddr  string
		wantIP   string
		wantPort uint16
		wantErr  bool
	}{
		{"0100007F:1F90", "127.0.0.1", 8080, false},
		{"00000000:0050", "0.0.0.0", 80, false},
		{"0100007F:0035", "127.0.0.1", 53, false},
		{"invalid", "", 0, true},
		{"0100007F", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.hexAddr, func(t *testing.T) {
			ip, port, err := ParseHexAddress(tt.hexAddr)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseHexAddress(%q) expected error", tt.hexAddr)
				}
				return
			}
			if err != nil {
				t.Errorf("ParseHexAddress(%q) unexpected error: %v", tt.hexAddr, err)
				return
			}
			if ip != tt.wantIP {
				t.Errorf("ParseHexAddress(%q) ip = %q, want %q", tt.hexAddr, ip, tt.wantIP)
			}
			if port != tt.wantPort {
				t.Errorf("ParseHexAddress(%q) port = %d, want %d", tt.hexAddr, port, tt.wantPort)
			}
		})
	}
}

func TestParseProcNetLine(t *testing.T) {
	line := "  0: 0100007F:1F90 5C28D893:01BB 01 00000000:00000000 00:00000000 00000000  1000        0 54321 1 0000000000000000 100 0 0 10 0"
	info, err := ParseProcNetLine(line, "tcp")
	if err != nil {
		t.Fatalf("ParseProcNetLine() error = %v", err)
	}
	if info.Protocol != "tcp" {
		t.Errorf("protocol = %q, want tcp", info.Protocol)
	}
	if info.SrcIP != "127.0.0.1" {
		t.Errorf("srcIP = %q, want 127.0.0.1", info.SrcIP)
	}
	if info.SrcPort != 8080 {
		t.Errorf("srcPort = %d, want 8080", info.SrcPort)
	}
	if info.DstPort != 443 {
		t.Errorf("dstPort = %d, want 443", info.DstPort)
	}
	if info.State != "ESTABLISHED" {
		t.Errorf("state = %q, want ESTABLISHED", info.State)
	}
	if info.Inode != 54321 {
		t.Errorf("inode = %d, want 54321", info.Inode)
	}
}

func TestBuildPortSet(t *testing.T) {
	ports := []int{1080, 9050, -1, 0, 70000}
	s := BuildPortSet(ports)
	if len(s) != 2 {
		t.Errorf("BuildPortSet len = %d, want 2 (invalid ports skipped)", len(s))
	}
	if _, ok := s[1080]; !ok {
		t.Error("port 1080 should be in set")
	}
	if _, ok := s[9050]; !ok {
		t.Error("port 9050 should be in set")
	}
}

func TestClassifyNetOp(t *testing.T) {
	socksPorts := BuildPortSet([]int{1080, 9050})

	tests := []struct {
		name        string
		info        types.NetConnInfo
		dnsTrace    bool
		socks5Ports map[uint16]struct{}
		want        types.NetOp
	}{
		{
			name:        "SOCKS5 connection",
			info:        types.NetConnInfo{Protocol: "tcp", DstPort: 1080},
			dnsTrace:    false,
			socks5Ports: socksPorts,
			want:        types.OpSOCKS5Connect,
		},
		{
			name:        "DNS query",
			info:        types.NetConnInfo{Protocol: "udp", DstPort: 53},
			dnsTrace:    true,
			socks5Ports: socksPorts,
			want:        types.OpDNSQuery,
		},
		{
			name:        "DNS query disabled",
			info:        types.NetConnInfo{Protocol: "udp", DstPort: 53},
			dnsTrace:    false,
			socks5Ports: socksPorts,
			want:        types.OpUDPConnect,
		},
		{
			name:        "UDP connection",
			info:        types.NetConnInfo{Protocol: "udp", DstPort: 8080},
			dnsTrace:    false,
			socks5Ports: socksPorts,
			want:        types.OpUDPConnect,
		},
		{
			name:        "TCP connection",
			info:        types.NetConnInfo{Protocol: "tcp", DstPort: 443},
			dnsTrace:    false,
			socks5Ports: socksPorts,
			want:        types.OpTCPConnect,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyNetOp(tt.info, tt.dnsTrace, tt.socks5Ports)
			if got != tt.want {
				t.Errorf("ClassifyNetOp() = %v, want %v", got, tt.want)
			}
		})
	}
}
