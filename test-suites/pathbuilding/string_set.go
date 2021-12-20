package pathbuilding

type StringSet struct {
	vals map[string]bool
}

func NewStringSet() *StringSet {
	return &StringSet{make(map[string]bool)}
}

func (ss *StringSet) Add(s string) {
	ss.vals[s] = true
}

func (ss *StringSet) Contains(s string) bool {
	_, ok := ss.vals[s]
	return ok
}

func (ss *StringSet) Remove(s string) {
	delete(ss.vals, s)
}

func (ss *StringSet) Values() []string {
	res := make([]string, 0, len(ss.vals))
	for k := range ss.vals {
		res = append(res, k)
	}
	return res
}
