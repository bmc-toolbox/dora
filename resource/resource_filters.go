package resource

type Filter struct {
	filters map[string][]string
}

func NewFilter() (f *Filter) {
	f = &Filter{filters: make(map[string][]string)}
	return f
}

func (f *Filter) Add(name string, values []string) {
	f.filters[name] = values
}

func (f *Filter) Get() map[string][]string {
	return f.filters
}

func (f *Filter) Clean() {
	f.filters = make(map[string][]string)
}
