package epochs

const (
	DayEpochIdentifier  = "day"
	WeekEpochIdentifier = "week"
)

var OsmosisIndexableEpochs = map[string]bool{
	DayEpochIdentifier:  true,
	WeekEpochIdentifier: true,
}
