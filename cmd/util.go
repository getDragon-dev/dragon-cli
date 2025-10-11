package cmd

import (
	"strconv"
	"strings"
)

type ver struct{ M, m, p int }

func parseVer(v string) ver {
	v = strings.TrimSpace(strings.TrimPrefix(v, "v"))
	parts := strings.SplitN(v, ".", 3)
	to := func(s string) int { n, _ := strconv.Atoi(s); return n }
	var out ver
	if len(parts) > 0 {
		out.M = to(parts[0])
	}
	if len(parts) > 1 {
		out.m = to(parts[1])
	}
	if len(parts) > 2 {
		out.p = to(parts[2])
	}
	return out
}

func cmpVer(a, b ver) int {
	if a.M != b.M {
		if a.M < b.M {
			return -1
		}
		return 1
	}
	if a.m != b.m {
		if a.m < b.m {
			return -1
		}
		return 1
	}
	if a.p != b.p {
		if a.p < b.p {
			return -1
		}
		return 1
	}
	return 0
}

type bounds struct {
	lower       *ver
	upper       *ver
	includeLow  bool
	includeHigh bool
}

// derive bounds for operators: >=, <=, >, <, =, ^, ~
func boundsFrom(op string, t ver) bounds {
	switch op {
	case ">=":
		return bounds{lower: &t, includeLow: true}
	case "<=":
		return bounds{upper: &t, includeHigh: true}
	case ">":
		return bounds{lower: &t}
	case "<":
		return bounds{upper: &t}
	case "=", "":
		return bounds{lower: &t, upper: &t, includeLow: true, includeHigh: true}
	case "^": // >= t, < (t.M+1).0.0
		u := ver{M: t.M + 1}
		return bounds{lower: &t, includeLow: true, upper: &u}
	case "~": // >= t, < t.M.(t.m+1).0
		u := ver{M: t.M, m: t.m + 1}
		return bounds{lower: &t, includeLow: true, upper: &u}
	default:
		// treat unknown operator as equality
		return bounds{lower: &t, upper: &t, includeLow: true, includeHigh: true}
	}
}

func satisfies(v, constraint string) bool {
	if constraint == "" {
		return true
	}
	// split op and version
	op, verStr := "", constraint
	for _, o := range []string{">=", "<=", ">", "<", "=", "^", "~"} {
		if strings.HasPrefix(constraint, o) {
			op = o
			verStr = strings.TrimSpace(strings.TrimPrefix(constraint, o))
			break
		}
	}

	val := parseVer(v)
	tgt := parseVer(verStr)
	b := boundsFrom(op, tgt)

	if b.lower != nil {
		c := cmpVer(val, *b.lower)
		if b.includeLow {
			if c < 0 {
				return false
			}
		} else {
			if c <= 0 {
				return false
			}
		}
	}
	if b.upper != nil {
		c := cmpVer(val, *b.upper)
		if b.includeHigh {
			if c > 0 {
				return false
			}
		} else {
			if c >= 0 {
				return false
			}
		}
	}
	return true
}
