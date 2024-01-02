package filter

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/DefiantLabs/cosmos-indexer/db/models"
)

type EventData struct {
	Event      models.BlockEvent
	Attributes []models.BlockEventAttribute
}

type BlockEventFilter interface {
	EventMatches(EventData) (bool, error)
	IncludeMatch() bool
	Valid() (bool, error)
}

type DefaultBlockEventTypeFilter struct {
	EventType string `json:"event_type"`
	Inclusive bool   `json:"inclusive"`
}

func (f DefaultBlockEventTypeFilter) EventMatches(eventData EventData) (bool, error) {
	return eventData.Event.BlockEventType.Type == f.EventType, nil
}

func (f DefaultBlockEventTypeFilter) IncludeMatch() bool {
	return f.Inclusive
}

func (f DefaultBlockEventTypeFilter) Valid() (bool, error) {
	if f.EventType != "" {
		return true, nil
	}

	return false, errors.New("EventType must be set")
}

type RegexBlockEventTypeFilter struct {
	EventTypeRegexPattern string `json:"event_type_regex"`
	eventTypeRegex        *regexp.Regexp
	Inclusive             bool `json:"inclusive"`
}

func (f RegexBlockEventTypeFilter) EventMatches(eventData EventData) (bool, error) {
	return f.eventTypeRegex.MatchString(eventData.Event.BlockEventType.Type), nil
}

func (f RegexBlockEventTypeFilter) IncludeMatch() bool {
	return f.Inclusive
}

func (f RegexBlockEventTypeFilter) Valid() (bool, error) {
	if f.eventTypeRegex != nil && f.EventTypeRegexPattern != "" {
		return true, nil
	}

	return false, errors.New("EventTypeRegexPattern must be set")
}

type DefaultBlockEventTypeAndAttributeValueFilter struct {
	EventType      string `json:"event_type"`
	AttributeKey   string `json:"attribute_key"`
	AttributeValue string `json:"attribute_value"`
	Inclusive      bool   `json:"inclusive"`
}

func (f DefaultBlockEventTypeAndAttributeValueFilter) EventMatches(eventData EventData) (bool, error) {
	if eventData.Event.BlockEventType.Type != f.EventType {
		return false, nil
	}

	for _, attr := range eventData.Attributes {
		if attr.BlockEventAttributeKey.Key == f.AttributeKey && attr.Value == f.AttributeValue {
			return true, nil
		}
	}

	return false, nil
}

func (f DefaultBlockEventTypeAndAttributeValueFilter) IncludeMatch() bool {
	return f.Inclusive
}

func (f DefaultBlockEventTypeAndAttributeValueFilter) Valid() (bool, error) {
	if f.EventType != "" && f.AttributeKey != "" && f.AttributeValue != "" {
		return true, nil
	}

	return false, errors.New("EventType, AttributeKey and AttributeValue must be set")
}

type RollingWindowBlockEventFilter interface {
	EventsMatch([]EventData) (bool, error)
	RollingWindowLength() int
	IncludeMatches() bool
	Valid() (bool, error)
}

type DefaultRollingWindowBlockEventFilter struct {
	EventPatterns  []BlockEventFilter
	includeMatches bool
}

func (f DefaultRollingWindowBlockEventFilter) EventsMatch(eventData []EventData) (bool, error) {
	if len(eventData) < f.RollingWindowLength() {
		return false, nil
	}

	for i, pattern := range f.EventPatterns {
		patternMatches, err := pattern.EventMatches(eventData[i])
		if !patternMatches || err != nil {
			return false, err
		}
	}

	return true, nil
}

func (f DefaultRollingWindowBlockEventFilter) IncludeMatches() bool {
	return f.includeMatches
}

func (f DefaultRollingWindowBlockEventFilter) RollingWindowLength() int {
	return len(f.EventPatterns)
}

func (f DefaultRollingWindowBlockEventFilter) Valid() (bool, error) {
	if len(f.EventPatterns) == 0 {
		return false, errors.New("eventPatterns must be set")
	}

	for index, pattern := range f.EventPatterns {
		valid, err := pattern.Valid()
		if !valid || err != nil {
			return false, fmt.Errorf("error parsing eventPatterns at index %d: %s", index, err)
		}
	}

	return true, nil
}

func NewDefaultBlockEventTypeFilter(eventType string, inclusive bool) BlockEventFilter {
	return &DefaultBlockEventTypeFilter{EventType: eventType, Inclusive: inclusive}
}

func NewDefaultBlockEventTypeAndAttributeValueFilter(eventType string, attributeKey string, attributeValue string, inclusive bool) BlockEventFilter {
	return &DefaultBlockEventTypeAndAttributeValueFilter{EventType: eventType, AttributeKey: attributeKey, AttributeValue: attributeValue, Inclusive: inclusive}
}

func NewRegexBlockEventFilter(eventTypeRegex string, inclusive bool) (BlockEventFilter, error) {
	re, err := regexp.Compile(eventTypeRegex)
	if err != nil {
		return nil, err
	}
	return &RegexBlockEventTypeFilter{EventTypeRegexPattern: eventTypeRegex, eventTypeRegex: re, Inclusive: inclusive}, nil
}

func NewDefaultRollingWindowBlockEventFilter(eventPatterns []BlockEventFilter, includeMatches bool) RollingWindowBlockEventFilter {
	return &DefaultRollingWindowBlockEventFilter{EventPatterns: eventPatterns, includeMatches: includeMatches}
}
