package int_set

import (
	"fmt"
	"strconv"
	"strings"
)

type IntRange struct {
	// Start and end are inclusive
	start int
	end   int
}
type IntSet struct {
	ranges []IntRange
}

func (is *IntSet) Set(s string) error {
	parts := strings.Split(s, ",")
	ranges := make([]IntRange, 0, len(parts))
	for _, part := range parts {
		r, err := parseIntRange(part)
		if err != nil {
			return err
		}
		ranges = append(ranges, r)
	}
	is.ranges = ranges
	return nil
}

func (is *IntSet) String() string {
	ranges := make([]string, 0, len(is.ranges))
	for _, r := range is.ranges {
		if r.start == r.end {
			ranges = append(ranges, strconv.Itoa(r.start))
		} else {
			ranges = append(ranges, strconv.Itoa(r.start)+"-"+strconv.Itoa(r.end))
		}
	}
	return strings.Join(ranges, ",")
}

func parseIntRange(s string) (IntRange, error) {
	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return IntRange{}, err
		}
		return IntRange{start, start}, nil
	}
	if len(parts) != 2 {
		return IntRange{}, fmt.Errorf("invalid integer range: %s", s)
	}
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return IntRange{}, err
	}
	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return IntRange{}, err
	}
	return IntRange{start, end}, nil
}

func (s *IntSet) Empty() bool {
	return len(s.ranges) == 0
}

func (s *IntSet) Contains(v int) bool {
	for _, r := range s.ranges {
		if v >= r.start && v <= r.end {
			return true
		}
	}
	return false
}
