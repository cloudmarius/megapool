package megapool

import (
	"errors"
	"math"
	"net/netip"
	"slices"
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
	var rangePool []Range
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
			rangePool = append(rangePool, r)
			continue
		}
		return Megapool{}, errors.New("not an IP, CIDR block or IP range")
	}
	return Megapool{
		IPPool:     ipPool,
		PrefixPool: prefixPool,
		RangePool:  rangePool,
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
	var ips1 []string
	var ips2 []string
	for _, v := range m.IPPool {
		ips1 = append(ips1, v.String())
	}
	for _, v := range other.IPPool {
		ips2 = append(ips2, v.String())
	}
	sort.Strings(ips1)
	sort.Strings(ips2)
	if !slices.Equal(ips1, ips2) {
		return false
	}

	var prefixes1 []string
	var prefixes2 []string
	for _, v := range m.PrefixPool {
		prefixes1 = append(prefixes1, v.String())
	}
	for _, v := range other.PrefixPool {
		prefixes2 = append(prefixes2, v.String())
	}
	sort.Strings(prefixes1)
	sort.Strings(prefixes2)
	if !slices.Equal(prefixes1, prefixes2) {
		return false
	}

	var ranges1 []string
	var ranges2 []string
	for _, v := range m.RangePool {
		ranges1 = append(ranges1, v.String())
	}
	for _, v := range other.RangePool {
		ranges2 = append(ranges2, v.String())
	}
	sort.Strings(ranges1)
	sort.Strings(ranges2)
	return slices.Equal(ranges1, ranges2)
}

func (m *Megapool) String() string {
	return strings.Join(m.AsSlice(), ",")
}

func (m *Megapool) AsSlice() []string {
	var s []string
	for _, v := range m.IPPool {
		s = append(s, v.String())
	}
	for _, v := range m.PrefixPool {
		s = append(s, v.String())
	}
	for _, v := range m.RangePool {
		s = append(s, v.String())
	}
	return s
}

func (r *Range) String() string {
	return r.From.String() + "-" + r.To.String()
}
