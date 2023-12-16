package filter

type StaticBlockEventFilterRegistry struct {
	BlockEventFilters         []BlockEventFilter
	RollingWindowEventFilters []RollingWindowBlockEventFilter
}

func (r *StaticBlockEventFilterRegistry) RegisterBlockEventFilter(filter BlockEventFilter) {
	r.BlockEventFilters = append(r.BlockEventFilters, filter)
}

func (r *StaticBlockEventFilterRegistry) RegisterRollingWindowBlockEventFilter(filter RollingWindowBlockEventFilter) {
	r.RollingWindowEventFilters = append(r.RollingWindowEventFilters, filter)
}

func (r *StaticBlockEventFilterRegistry) NumFilters() int {
	return len(r.BlockEventFilters) + len(r.RollingWindowEventFilters)
}
