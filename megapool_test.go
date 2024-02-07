package megapool

import (
	"net/netip"
	"testing"
)

func TestNewMegapool(t *testing.T) {
	type args struct {
		ipsAndCIDRs string
	}
	tests := []struct {
		name    string
		args    args
		want    Megapool
		wantErr bool
	}{
		{
			"empty",
			args{""},
			Megapool{nil, nil},
			false,
		}, {
			"spaces",
			args{"   		"},
			Megapool{nil, nil},
			false,
		}, {
			"wrong IP field missing",
			args{"8.8.8/32"},
			Megapool{nil, nil},
			true,
		}, {
			"wrong IP field missing at least one digit",
			args{"8.8.8./32"},
			Megapool{nil, nil},
			true,
		}, {
			"wrong IP field out of range >255",
			args{"8.8.8.888/32"},
			Megapool{nil, nil},
			true,
		}, {
			"wrong dash separator",
			args{"8.8.8.8-1.1.1.1"},
			Megapool{nil, nil},
			true,
		}, {
			"wrong CIDR field out of range >255",
			args{"8.8.8.888/8"},
			Megapool{nil, nil},
			true,
		}, {
			"wrong CIDR prefix of range >32",
			args{"8.8.8.8/88"},
			Megapool{nil, nil},
			true,
		}, {
			"only IPs and comma separator and ordered and spaces and tabs",
			args{"8.8.8.7,8.8.8.8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				nil},
			false,
		}, {
			"only CIDRs and comma separator and ordered and spaces and tabs",
			args{"1.1.1.1/8,2.2.2.2/8"},
			Megapool{
				nil,
				[]netip.Prefix{p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		}, {
			"comma separator and ordered and spaces and tabs",
			args{"8.8.8.7,1.1.1.1/8,8.8.8.8, 2.2.2.2/8,		3.3.3.3/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("1.1.1.1/8"), p("2.2.2.2/8"), p("3.3.3.3/8")}},
			false,
		}, {
			"comma separator and unordered and spaces and tabs",
			args{"8.8.8.8,8.8.8.7,1.1.1.1/8, 2.2.2.2/8,		3.3.3.3/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.3.3.3/8"), p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		}, {
			"semicolon separator and unordered and spaces and tabs",
			args{"8.8.8.8;8.8.8.7;1.1.1.1/8; 2.2.2.2/8;		3.3.3.3/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.3.3.3/8"), p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		}, {
			"new line separator and unordered and spaces and tabs",
			args{`8.8.8.8
8.8.8.7
1.1.1.1/8
 2.2.2.2/8
		3.3.3.3/8`},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.3.3.3/8"), p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		}, {
			"escaped new line separator and unordered and spaces and tabs",
			args{"8.8.8.8\n8.8.8.7\n1.1.1.1/8\n2.2.2.2/8\n\t3.3.3.3/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.3.3.3/8"), p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		}, {
			"mixed separators and unordered and spaces and tabs",
			args{"8.8.8.8,8.8.8.7;1.1.1.1/8, 2.2.2.2/8;		3.3.3.3/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.3.3.3/8"), p("1.1.1.1/8"), p("2.2.2.2/8")}},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMegapool(tt.args.ipsAndCIDRs)
			//fmt.Println(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMegapool() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("NewMegapool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegapool_Overlaps(t *testing.T) {
	tests := []struct {
		name string
		main string
		args string
		want bool
	}{
		{"only CIDRs and overlapping and unordered", "2.2.2.2/8,1.1.1.1/8", "1.1.1.1/24,3.3.3.3/24", true},
		{"only CIDRs and not overlapping and unordered", "2.2.2.2/8,1.1.1.1/8", "4.4.4.4/8,3.3.3.3/8", false},
		{"only IPs and overlapping and unordered", "1.1.1.1,2.2.2.2", "3.3.3.3,1.1.1.1", true},
		{"only IPs and not overlapping and unordered", "1.1.1.1,2.2.2.2", "3.3.3.3,4.4.4.4", false},
		{"mixed and overlapping IP left and unordered", "3.3.3.255,2.2.2.2/8,1.1.1.1/8", "4.4.4.4/8,3.3.3.3/8", true},
		{"mixed and overlapping IP right and unordered", "2.2.2.2/8,1.1.1.1/8", "1.1.1.255,4.4.4.4/8,3.3.3.3/8", true},
		{"mixed and overlapping IP right and left and unordered", "5.5.5.5,2.2.2.2/8,1.1.1.1/8", "5.5.5.5,4.4.4.4/8,3.3.3.3/8", true},
		{"mixed and not overlapping", "5.5.5.5,2.2.2.2/8,1.1.1.1/8", "6.6.6.6,4.4.4.4/8,3.3.3.3/8", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMegapool(tt.main)
			other, _ := NewMegapool(tt.args)
			if got := m.Overlaps(other); got != tt.want {
				t.Errorf("Megapool.Overlaps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegapool_HasMinSize(t *testing.T) {
	tests := []struct {
		name string
		main string
		args int
		want bool
	}{
		{"empty", "", 1, false},
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 3, true},
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 4, false},
		{"only CIDRs", "1.1.1.1/32", 2, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/30", 10, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/29", 10, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/28", 10, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/28", 17, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/28", 18, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/24", 257, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/24", 258, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/16", 65537, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/16", 65538, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/8", 16777217, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/8", 16777218, false},
		{"mixed IPs and CIDRs", "1.1.1.1,1.1.1.2,1.2.1.1/24,1.3.1.1/24", 514, true},
		{"mixed IPs and CIDRs", "1.1.1.1,1.1.1.2,1.2.1.1/24,1.3.1.1/24", 515, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMegapool(tt.main)
			if got := m.HasMinSize(tt.args); got != tt.want {
				t.Errorf("Megapool.HasMinSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func p(s string) netip.Prefix {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return p
}

func a(s string) netip.Addr {
	a, err := netip.ParseAddr(s)
	if err != nil {
		panic(err)
	}
	return a
}