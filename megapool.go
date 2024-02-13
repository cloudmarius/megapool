package megapool

import (
	"errors"
	"math"
	"net/netip"
	"reflect"
	"sort"
	"strings"
)

type Megapool struct {
	IPPool     []netip.Addr
	PrefixPool []netip.Prefix
	RangePool  []Range
}

type Range struct {
	From netip.Addr
	To   netip.Addr
}

func NewMegapool(ipsAndCIDRs string) (Megapool, error) {
	var ipPool []netip.Addr
	var prefixPool []netip.Prefix
	var ranges []Range
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
			continue
		}
		p, err := netip.ParsePrefix(vv)
		if err == nil {
			prefixPool = append(prefixPool, p)
			continue
		}
		r, err := parseRange(vv)
		if err == nil {
			ranges = append(ranges, r)
			continue
		}
		return Megapool{}, errors.New("not an IP, CIDR block or IP range")
	}
	return Megapool{
		IPPool:     ipPool,
		PrefixPool: prefixPool,
		RangePool:  ranges,
	}, nil
}

func parseRange(r string) (Range, error) {
	items := strings.Split(r, "-")
	if len(items) != 2 {
		return Range{}, errors.New("not an accepted range")
	}

	from, err := netip.ParseAddr(items[0])
	if err != nil {
		return Range{}, errors.New("not an accepted range")
	}
	to, err := netip.ParseAddr(items[1])
	if err != nil {
		return Range{}, errors.New("not an accepted range")
	}
	fromSlice := from.AsSlice()
	toSlice := to.AsSlice()
	if len(fromSlice) == len(toSlice) {
		for i := 0; i < len(fromSlice)-1; i++ {
			if fromSlice[i] != toSlice[i] {
				return Range{}, errors.New("not an accepted range")
			}
		}
		if fromSlice[len(fromSlice)-1] >= toSlice[len(toSlice)-1] {
			return Range{}, errors.New("not an accepted range")
		}
	} else {
		return Range{}, errors.New("not an accepted range")
	}
	return Range{From: from, To: to}, nil
}

func (m *Megapool) Overlaps(others ...Megapool) bool {
	for _, o := range others {
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

		for _, p1 := range m.PrefixPool {
			for _, r2 := range o.RangePool {
				if p1.Contains(r2.From) || p1.Contains(r2.To) {
					return true
				}
			}
		}
		for _, p2 := range o.PrefixPool {
			for _, r1 := range m.RangePool {
				if p2.Contains(r1.From) || p2.Contains(r1.To) {
					return true
				}
			}
		}
		for _, r1 := range m.RangePool {
			for _, ip2 := range o.IPPool {
				if r1.From.Compare(ip2) <= 0 && r1.To.Compare(ip2) >= 0 {
					return true
				}
			}
		}
		for _, r2 := range o.RangePool {
			for _, ip1 := range m.IPPool {
				if r2.From.Compare(ip1) <= 0 && r2.To.Compare(ip1) >= 0 {
					return true
				}
			}
		}
		for _, r1 := range m.RangePool {
			for _, r2 := range o.RangePool {
				if (r1.From.Compare(r2.From) <= 0 && r1.To.Compare(r2.From) >= 0) ||
					(r1.From.Compare(r2.To) <= 0 && r1.To.Compare(r2.To) >= 0) {
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
	for _, v := range m.RangePool {
		from := v.From.AsSlice()
		to := v.To.AsSlice()
		if len(from) == 4 && len(to) == 4 {
			actual += float64(to[3] - from[3] + 1)
			if actual >= min {
				return true
			}
		}
	}
	return false
}

func (m *Megapool) HasMaxSize(maxSize int) bool {
	if maxSize == 0 {
		return true
	}
	max := float64(maxSize)
	actual := float64(len(m.IPPool))
	if actual > max {
		return false
	}
	for _, v := range m.PrefixPool {
		actual += math.Pow(2, float64(32-v.Bits()))
		if actual > max {
			return false
		}
	}
	for _, v := range m.RangePool {
		from := v.From.AsSlice()
		to := v.To.AsSlice()
		if len(from) == 4 && len(to) == 4 {
			actual += float64(to[3] - from[3] + 1)
			if actual > max {
				return false
			}
		}
	}
	return actual <= max
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
	if !reflect.DeepEqual(m1Prefixes, m2Prefixes) {
		return false
	}

	var m1Ranges []string
	var m2Ranges []string
	for _, v := range m.RangePool {
		m1Ranges = append(m1Ranges, v.String())
	}
	for _, v := range other.RangePool {
		m2Ranges = append(m2Ranges, v.String())
	}
	sort.Strings(m1Ranges)
	sort.Strings(m2Ranges)
	return reflect.DeepEqual(m1Ranges, m2Ranges)
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

func (r *Range) String() string {
	return r.From.String() + "-" + r.To.String()
}
