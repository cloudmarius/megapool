package megapool

import (
	"net/netip"
	"reflect"
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
			Megapool{nil, nil, nil},
			false,
		}, {
			"spaces",
			args{"   		"},
			Megapool{nil, nil, nil},
			false,
		}, {
			"wrong IP field missing",
			args{"8.8.8/32"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong IP field missing at least one digit",
			args{"8.8.8./32"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong IP field out of range >255",
			args{"8.8.8.888/32"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong separator",
			args{"8.8.8.8_1.1.1.1"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong CIDR field out of range >255",
			args{"8.8.8.888/8"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong CIDR prefix of range >32",
			args{"8.8.8.8/88"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong range not ordered",
			args{"8.8.8.8-8.8.8.7"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong range only last segment can be different",
			args{"8.8.8.8-8.8.80.10"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"wrong range bad ip",
			args{"8.8.8.8-8.8.8"},
			Megapool{nil, nil, nil},
			true,
		}, {
			"only IPs and comma separator and ordered",
			args{"8.8.8.7,8.8.8.8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				nil, nil},
			false,
		}, {
			"only CIDRs and comma separator and ordered",
			args{"1.0.0.0/8,2.0.0.0/8"},
			Megapool{
				nil,
				[]netip.Prefix{p("1.0.0.0/8"), p("2.0.0.0/8")},
				nil},
			false,
		}, {
			"only ranges and comma separator",
			args{"1.1.1.1-1.1.1.10,2.2.2.0-2.2.2.5"},
			Megapool{
				nil,
				nil,
				[]Range{{From: a("1.1.1.1"), To: a("1.1.1.10")}, {From: a("2.2.2.0"), To: a("2.2.2.5")}}},
			false,
		}, {
			"comma separator and ordered and spaces and tabs",
			args{"8.8.8.7,1.0.0.0/8,8.8.8.8, 2.0.0.0/8,		3.0.0.0/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("1.0.0.0/8"), p("2.0.0.0/8"), p("3.0.0.0/8")},
				nil},
			false,
		}, {
			"comma separator and unordered and spaces and tabs",
			args{"8.8.8.8,8.8.8.7,1.0.0.0/8, 2.0.0.0/8,		3.0.0.0/8,1.1.1.1-1.1.1.10,2.2.2.0-2.2.2.5"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.0.0.0/8"), p("1.0.0.0/8"), p("2.0.0.0/8")},
				[]Range{{From: a("1.1.1.1"), To: a("1.1.1.10")}, {From: a("2.2.2.0"), To: a("2.2.2.5")}},
			},
			false,
		}, {
			"semicolon separator and unordered and spaces and tabs",
			args{"8.8.8.8;8.8.8.7;1.0.0.0/8; 2.0.0.0/8;		3.0.0.0/8;1.1.1.1-1.1.1.10;2.2.2.0-2.2.2.5"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.0.0.0/8"), p("1.0.0.0/8"), p("2.0.0.0/8")},
				[]Range{{From: a("1.1.1.1"), To: a("1.1.1.10")}, {From: a("2.2.2.0"), To: a("2.2.2.5")}},
			},
			false,
		}, {
			"new line separator and unordered and spaces and tabs",
			args{`8.8.8.8
8.8.8.7
1.0.0.0/8
 2.0.0.0/8
		3.0.0.0/8
	1.1.1.1-1.1.1.10
2.2.2.0-2.2.2.5`},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.0.0.0/8"), p("1.0.0.0/8"), p("2.0.0.0/8")},
				[]Range{{From: a("1.1.1.1"), To: a("1.1.1.10")}, {From: a("2.2.2.0"), To: a("2.2.2.5")}},
			},
			false,
		}, {
			"escaped new line separator and unordered and spaces and tabs",
			args{"8.8.8.8\n8.8.8.7\n1.0.0.0/8\n2.0.0.0/8\n\t3.0.0.0/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				[]netip.Prefix{p("3.0.0.0/8"), p("1.0.0.0/8"), p("2.0.0.0/8")},
				nil},
			false,
		}, {
			"mixed separators and unordered and spaces and tabs",
			args{"8.8.8.8,8.8.8.7;1.0.0.0/8, 2.0.0.0/8;		3.0.0.0/8"},
			Megapool{
				[]netip.Addr{a("8.8.8.7"), a("8.8.8.8")},
				append([]netip.Prefix{p("3.0.0.0/8"), p("1.0.0.0/8"), p("2.0.0.0/8")}),
				nil},
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
		{"only CIDRs and overlapping and unordered", "2.0.0.0/8,1.0.0.0/8", "1.1.1.1/24,3.3.3.3/24", true},
		{"only CIDRs and not overlapping", "2.0.0.0/8,1.0.0.0/8", "4.0.0.0/8,3.0.0.0/8", false},
		{"only IPs and overlapping and unordered", "1.1.1.1,2.2.2.2", "3.3.3.3,1.1.1.1", true},
		{"only IPs and not overlapping", "1.1.1.1,2.2.2.2", "3.3.3.3,4.4.4.4", false},
		{"only ranges and overlapping contained", "1.1.1.2-1.1.1.10", "1.1.1.4-1.1.1.6", true},
		{"only ranges and overlapping intersect right", "1.1.1.2-1.1.1.10", "1.1.1.4-1.1.1.12", true},
		{"only ranges and overlapping intersect right just one", "1.1.1.2-1.1.1.10", "1.1.1.10-1.1.1.12", true},
		{"only ranges and overlapping intersect left", "1.1.1.2-1.1.1.10", "1.1.1.1-1.1.1.3", true},
		{"only ranges and overlapping intersect left just one", "1.1.1.2-1.1.1.10", "1.1.1.1-1.1.1.2", true},
		{"only IPs and ranges and overlapping", "1.1.1.2-1.1.1.10", "1.1.1.2", true},
		{"only IPs and ranges and overlapping", "1.1.1.2-1.1.1.10", "1.1.1.5", true},
		{"only IPs and ranges and overlapping", "1.1.1.2-1.1.1.10", "1.1.1.10", true},
		{"only IPs and ranges and overlapping", "1.1.1.5", "1.1.1.2-1.1.1.10", true},
		{"only IPs and ranges and overlapping", "1.1.1.2", "1.1.1.2-1.1.1.10", true},
		{"only IPs and ranges and overlapping", "1.1.1.10", "1.1.1.2-1.1.1.10", true},
		{"only IPs and ranges and not overlapping left", "1.1.1.2-1.1.1.10", "1.1.1.1", false},
		{"only IPs and ranges and not overlapping right", "1.1.1.2-1.1.1.10", "1.1.1.11", false},
		{"only IPs and ranges and not overlapping left", "1.1.1.1", "1.1.1.2-1.1.1.10", false},
		{"only IPs and ranges and not overlapping right", "1.1.1.11", "1.1.1.2-1.1.1.10", false},
		{"only CIDRs and ranges and overlapping", "1.0.0.0/8", "1.1.1.1-1.1.1.10", true},
		{"only CIDRs and ranges and overlapping", "1.1.1.0/24", "1.1.1.1-1.1.1.10", true},
		{"only CIDRs and ranges and overlapping", "1.1.1.1-1.1.1.10", "1.0.0.0/8", true},
		{"only CIDRs and ranges and overlapping", "1.1.1.1-1.1.1.10", "1.1.1.0/24", true},
		{"mixed and overlapping IP left and unordered", "3.3.3.255,2.0.0.0/8,1.0.0.0/8,10.10.10.10-10.10.10.17", "4.0.0.0/8,3.0.0.0/8", true},
		{"mixed and overlapping IP left and unordered", "3.3.3.255,2.0.0.0/8,1.0.0.0/8,10.10.10.10-10.10.10.17", "4.0.0.0/8,3.3.3.250-3.3.3.255", true},
		{"mixed and overlapping IP right and unordered", "2.0.0.0/8,1.0.0.0/8", "1.1.1.255,4.0.0.0/8,3.0.0.0/8", true},
		{"mixed and overlapping IP right and unordered", "2.0.0.0/8,1.1.1.250-1.1.1.255", "1.1.1.255,4.0.0.0/8,3.0.0.0/8", true},
		{"mixed and overlapping IP right and left and unordered", "5.5.5.5,2.0.0.0/8,1.0.0.0/8", "5.5.5.5,4.0.0.0/8,3.0.0.0/8", true},
		{"mixed and not overlapping", "5.5.5.5,2.0.0.0/8,1.0.0.0/8,6.6.6.1-6.6.6.5", "6.6.6.6,4.0.0.0/8,3.0.0.0/8,5.5.5.1-5.5.5.2", false},
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
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 2, true},
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
		{"only ranges and less", "1.1.1.1-1.1.1.10", 9, true},
		{"only ranges and equal", "1.1.1.1-1.1.1.10", 10, true},
		{"only ranges too much", "1.1.1.1-1.1.1.10", 11, false},
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

func TestMegapool_HasMaxSize(t *testing.T) {
	tests := []struct {
		name string
		main string
		args int
		want bool
	}{
		{"empty", "", 0, true},
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 2, false},
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 3, true},
		{"only 3 IPs", "1.1.1.1,1.1.1.3,1.1.1.3", 4, true},
		{"only CIDRs /24", "1.1.1.1/24", 256, true},
		{"only CIDRs /32", "1.1.1.1/32", 2, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/30", 4, false},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/30", 5, true},
		{"only CIDRs", "1.1.1.1/32,1.2.1.1/30", 10, true},
		{"only ranges", "1.1.1.1-1.1.1.10", 9, false},
		{"only ranges", "1.1.1.1-1.1.1.10", 10, true},
		{"only ranges", "1.1.1.1-1.1.1.10", 11, true},
		{"only ranges and less", "1.1.1.0-1.1.1.10", 10, false},
		{"only ranges and equal", "1.1.1.0-1.1.1.10", 11, true},
		{"only ranges too much", "1.1.1.0-1.1.1.10", 12, true},
		{"mixed and less", "1.1.1.1,1.1.1.11-1.1.1.15,1.2.1.0/24", 261, false},
		{"mixed and match", "1.1.1.1,1.1.1.11-1.1.1.15,1.2.1.0/24", 262, true},
		{"mixed and more", "1.1.1.1,1.1.1.11-1.1.1.15,1.2.1.0/24", 263, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMegapool(tt.main)
			if got := m.HasMaxSize(tt.args); got != tt.want {
				t.Errorf("Megapool.HasMaxSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMegapool_AsSlice(t *testing.T) {
	tests := []struct {
		name string
		args string
		want []string
	}{
		{"empty", "", nil},
		{"shuffled",
			"1.1.1.1,1.1.1.5-1.1.1.10,1.1.1.2,2.2.2.0/24,1.1.1.20-1.1.1.25,2.2.3.0/24",
			[]string{"1.1.1.1", "1.1.1.2", "2.2.2.0/24", "2.2.3.0/24", "1.1.1.5-1.1.1.10", "1.1.1.20-1.1.1.25"}},
		{"shuffled some more",
			"2.2.2.0/24,1.1.1.5-1.1.1.10,1.1.1.1,1.1.1.20-1.1.1.25,2.2.3.0/24,1.1.1.2,",
			[]string{"1.1.1.1", "1.1.1.2", "2.2.2.0/24", "2.2.3.0/24", "1.1.1.5-1.1.1.10", "1.1.1.20-1.1.1.25"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := NewMegapool(tt.args)
			if got := m.AsSlice(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Megapool.AsSlice() = %v, want %v", got, tt.want)
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
