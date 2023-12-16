package filter

import (
	"github.com/DefiantLabs/cosmos-indexer/db/models"
)

type FilterEventData struct {
	Event      models.BlockEvent
	Attributes []models.BlockEventAttribute
}

type BlockEventFilter interface {
	EventMatches(FilterEventData) (bool, error)
	IncludeMatch() bool
}

type defaultBlockEventTypeFilter struct {
	eventType string
	inclusive bool
}

func (f *defaultBlockEventTypeFilter) EventMatches(eventData FilterEventData) (bool, error) {
	return eventData.Event.BlockEventType.Type == f.eventType, nil
}

func (f *defaultBlockEventTypeFilter) IncludeMatch() bool {
	return f.inclusive
}

type defaultBlockEventTypeAndAttributeValueFilter struct {
	eventType      string
	attributeKey   string
	attributeValue string
	inclusive      bool
}

func (f *defaultBlockEventTypeAndAttributeValueFilter) EventMatches(eventData FilterEventData) (bool, error) {
	if eventData.Event.BlockEventType.Type != f.eventType {
		return false, nil
	}

	for _, attr := range eventData.Attributes {
		if attr.BlockEventAttributeKey.Key == f.attributeKey && attr.Value == f.attributeValue {
			return true, nil
		}
	}

	return false, nil
}

func (f *defaultBlockEventTypeAndAttributeValueFilter) IncludeMatch() bool {
	return f.inclusive
}

type RollingWindowBlockEventFilter interface {
	EventsMatch([]FilterEventData) (bool, error)
	RollingWindowLength() int
	IncludeMatches() bool
}

type defaultRollingWindowBlockEventFilter struct {
	eventPatterns  []BlockEventFilter
	includeMatches bool
}

func (f *defaultRollingWindowBlockEventFilter) EventsMatch(eventData []FilterEventData) (bool, error) {
	if len(eventData) < f.RollingWindowLength() {
		return false, nil
	}

	for i, pattern := range f.eventPatterns {
		patternMatches, err := pattern.EventMatches(eventData[i])
		if !patternMatches || err != nil {
			return false, err
		}
	}

	return true, nil
}

func (f *defaultRollingWindowBlockEventFilter) RollingWindowLength() int {
	return len(f.eventPatterns)
}

func (f *defaultRollingWindowBlockEventFilter) IncludeMatches() bool {
	return f.includeMatches
}

func NewDefaultBlockEventTypeFilter(eventType string, inclusive bool) BlockEventFilter {
	return &defaultBlockEventTypeFilter{eventType: eventType, inclusive: inclusive}
}

func NewDefaultBlockEventTypeAndAttributeValueFilter(eventType string, attributeKey string, attributeValue string, inclusive bool) BlockEventFilter {
	return &defaultBlockEventTypeAndAttributeValueFilter{eventType: eventType, attributeKey: attributeKey, attributeValue: attributeValue, inclusive: inclusive}
}

func NewDefaultRollingWindowBlockEventFilter(eventPatterns []BlockEventFilter, includeMatches bool) RollingWindowBlockEventFilter {
	return &defaultRollingWindowBlockEventFilter{eventPatterns: eventPatterns, includeMatches: includeMatches}
}
