package resource

// Filter is meant to store the filters of requested via api
type Filter struct {
	filters map[string][]string
}

// NewFilter returns an empty new filter structure
func NewFilter() (f *Filter) {
	f = &Filter{filters: make(map[string][]string)}
	return f
}

// Add adds a new filter to the filter map
func (f *Filter) Add(name string, values []string) {
	f.filters[name] = values
}

// Get retrieve all filters
func (f *Filter) Get() map[string][]string {
	return f.filters
}

// Clean cleanup the current filter list
func (f *Filter) Clean() {
	f.filters = make(map[string][]string)
}

type byInt64Slice []int

func (b byInt64Slice) Len() int           { return len(b) }
func (b byInt64Slice) Swap(x, y int)      { b[x], b[y] = b[y], b[x] }
func (b byInt64Slice) Less(x, y int) bool { return b[x] < b[y] }
