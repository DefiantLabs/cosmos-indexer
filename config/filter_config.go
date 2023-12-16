package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/filter"
)

type blockEventFilterConfigs struct {
	BeginBlockFilters []json.RawMessage `json:"begin_block_filters"`
	EndBlockFilters   []json.RawMessage `json:"end_block_filters"`
}

type BlockEventFilterConfig struct {
	Type       string            `json:"type"`
	Subfilters []json.RawMessage `json:"subfilters"`
	Inclusive  bool              `json:"inclusive"`
}

func ParseJsonFilterConfig(configJson []byte) ([]filter.BlockEventFilter, []filter.RollingWindowBlockEventFilter, []filter.BlockEventFilter, []filter.RollingWindowBlockEventFilter, error) {
	config := blockEventFilterConfigs{}
	err := json.Unmarshal(configJson, &config)

	if err != nil {
		return nil, nil, nil, nil, err
	}

	beginBlockSingleEventFilters, beginBlockRollingWindowFilters, err := ParseLifecycleConfig(config.BeginBlockFilters)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	endBlockSingleEventFilters, endBlockRollingWindowFilters, err := ParseLifecycleConfig(config.EndBlockFilters)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	return beginBlockSingleEventFilters, beginBlockRollingWindowFilters, endBlockSingleEventFilters, endBlockRollingWindowFilters, nil
}

func ParseLifecycleConfig(lifecycleConfig []json.RawMessage) ([]filter.BlockEventFilter, []filter.RollingWindowBlockEventFilter, error) {
	rollingWindowFilters := []filter.RollingWindowBlockEventFilter{}
	singleEventFilters := []filter.BlockEventFilter{}
	for index, beginFilters := range lifecycleConfig {

		newFilter := BlockEventFilterConfig{}

		err := json.Unmarshal(beginFilters, &newFilter)

		if err != nil {
			parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
			return nil, nil, parserError
		}

		err = validateBlockEventFilterConfig(newFilter)
		if err != nil {
			parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
			return nil, nil, parserError
		}

		switch newFilter.Type {
		case "rolling_window":
			eventPatterns := []filter.BlockEventFilter{}
			for _, subfilter := range newFilter.Subfilters {
				newSubFilter := BlockEventFilterConfig{}
				err := json.Unmarshal(subfilter, &newSubFilter)
				if err != nil {
					parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
					return nil, nil, parserError
				}
				parsedFilter, err := ParseJsonFilterConfigFromType(newSubFilter.Type, subfilter, 0)
				if err != nil {
					parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
					return nil, nil, parserError
				}
				eventPatterns = append(eventPatterns, parsedFilter)
			}
			newRollingFilter := filter.NewDefaultRollingWindowBlockEventFilter(eventPatterns, newFilter.Inclusive)
			valid, err := newRollingFilter.Valid()
			if !valid || err != nil {
				parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
				return nil, nil, parserError
			}
			rollingWindowFilters = append(rollingWindowFilters, newRollingFilter)
		case "event_type":
			parsedFilter, err := ParseJsonFilterConfigFromType(newFilter.Type, beginFilters, 0)
			if err != nil {
				parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
				return nil, nil, parserError
			}
			valid, err := parsedFilter.Valid()
			if !valid || err != nil {
				parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
				return nil, nil, parserError
			}
			singleEventFilters = append(singleEventFilters, parsedFilter)
		case "event_type_and_attribute_value":
			parsedFilter, err := ParseJsonFilterConfigFromType(newFilter.Type, beginFilters, 0)
			if err != nil {
				parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
				return nil, nil, parserError
			}
			valid, err := parsedFilter.Valid()
			if !valid || err != nil {
				parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
				return nil, nil, parserError
			}
			singleEventFilters = append(singleEventFilters, parsedFilter)
		default:
			parserError := fmt.Errorf("error parsing begin_block_filters at index %d: unknown filter type \"%s\"", index, newFilter.Type)
			return nil, nil, parserError
		}

		if err != nil {
			parserError := fmt.Errorf("error parsing begin_block_filters at index %d: %s", index, err)
			return nil, nil, parserError
		}
	}

	return singleEventFilters, rollingWindowFilters, nil
}

func ParseJsonFilterConfigFromType(filterType string, configJson []byte, level int) (filter.BlockEventFilter, error) {
	switch filterType {
	case "event_type":
		newFilter := filter.DefaultBlockEventTypeFilter{}

		err := json.Unmarshal(configJson, &newFilter)

		if err != nil {
			return nil, err
		}
		return newFilter, nil
	case "event_type_and_attribute_value":
		newFilter := filter.DefaultBlockEventTypeAndAttributeValueFilter{}

		err := json.Unmarshal(configJson, &newFilter)

		if err != nil {
			return nil, err
		}
		return newFilter, nil
	default:
		return nil, fmt.Errorf("unknown filter type %s", filterType)
	}
}

func validateBlockEventFilterConfig(config BlockEventFilterConfig) error {
	if config.Type == "" {
		return errors.New("filter config must have a type field")
	}
	return nil
}
