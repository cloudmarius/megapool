package megapool

import (
	"math"
	"net/netip"
	"reflect"
	"sort"
	"strings"
)

type Megapool struct {
	IPPool     []netip.Addr
	PrefixPool []netip.Prefix
}

func NewMegapool(ipsAndCIDRs string) (Megapool, error) {
	var ipPool []netip.Addr
	var prefixPool []netip.Prefix
	items := strings.TrimSpace(ipsAndCIDRs)
	if len(items) == 0 {
		return Megapool{}, nil
	}
	all := strings.FieldsFunc(items, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
	for _, v := range all {
		vv := strings.ReplaceAll(strings.ReplaceAll(v, " ", ""), "\t", "")
		a, err := netip.ParseAddr(vv)
		if err == nil {
			ipPool = append(ipPool, a)
		} else {
			p, err := netip.ParsePrefix(vv)
			if err != nil {
				return Megapool{}, err
			}
			prefixPool = append(prefixPool, p)
		}
	}
	return Megapool{
		IPPool:     ipPool,
		PrefixPool: prefixPool,
	}, nil
}

func (m *Megapool) Overlaps(other ...Megapool) bool {
	for _, o := range other {
		for _, p1 := range m.PrefixPool {
			for _, p2 := range o.PrefixPool {
				if p1.Overlaps(p2) {
					return true
				}
			}
		}
		for _, p1 := range m.PrefixPool {
			for _, ip2 := range o.IPPool {
				if p1.Contains(ip2) {
					return true
				}
			}
		}
		for _, p2 := range o.PrefixPool {
			for _, ip1 := range m.IPPool {
				if p2.Contains(ip1) {
					return true
				}
			}
		}
		for _, ip1 := range m.IPPool {
			for _, ip2 := range o.IPPool {
				if ip1 == ip2 {
					return true
				}
			}
		}
	}
	return false
}

func (m *Megapool) HasMinSize(minSize int) bool {
	min := float64(minSize)
	actual := float64(len(m.IPPool))
	if actual >= min {
		return true
	}
	for _, v := range m.PrefixPool {
		actual += math.Pow(2, float64(32-v.Bits()))
		if actual >= min {
			return true
		}
	}
	return false
}

func (m *Megapool) HasMaxSize(maxSize int) bool {
	max := float64(maxSize)
	actual := float64(len(m.IPPool))
	if actual >= max {
		return false
	}
	for _, v := range m.PrefixPool {
		actual += math.Pow(2, float64(32-v.Bits()))
		if actual >= max {
			return false
		}
	}
	return true
}

func (m *Megapool) Equal(other Megapool) bool {
	var m1IPs []string
	var m2IPs []string
	for _, v := range m.IPPool {
		m1IPs = append(m1IPs, v.String())
	}
	for _, v := range other.IPPool {
		m2IPs = append(m2IPs, v.String())
	}
	sort.Strings(m1IPs)
	sort.Strings(m2IPs)
	if !reflect.DeepEqual(m1IPs, m2IPs) {
		return false
	}

	var m1Prefixes []string
	var m2Prefixes []string
	for _, v := range m.PrefixPool {
		m1Prefixes = append(m1Prefixes, v.String())
	}
	for _, v := range other.PrefixPool {
		m2Prefixes = append(m2Prefixes, v.String())
	}
	sort.Strings(m1Prefixes)
	sort.Strings(m2Prefixes)
	return reflect.DeepEqual(m1Prefixes, m2Prefixes)
}

func (m *Megapool) String() string {
	var all []string
	for _, v := range m.IPPool {
		all = append(all, v.String())
	}
	for _, v := range m.PrefixPool {
		all = append(all, v.String())
	}
	return strings.Join(all, ",")
}
