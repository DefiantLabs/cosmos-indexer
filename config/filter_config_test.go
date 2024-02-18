package config

import (
	"encoding/json"
	"testing"

	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/filter"
	"github.com/stretchr/testify/suite"
)

type FilterConfigTestSuite struct {
	suite.Suite
}

//nolint:dogsled
func (suite *FilterConfigTestSuite) TestParseJSONFilterConfig() {
	conf := blockFilterConfigs{}

	beginFilterEventTypeInvalid, err := getMockEventTypeBytes(true)

	suite.Require().NoError(err)

	conf.BeginBlockFilters = []json.RawMessage{beginFilterEventTypeInvalid}

	confBytes, err := json.Marshal(conf)
	suite.Require().NoError(err)

	_, _, _, _, _, err = ParseJSONFilterConfig(confBytes)

	suite.Require().Error(err)

	beginFilterEventTypeValid, err := getMockEventTypeBytes(false)
	suite.Require().NoError(err)

	conf.BeginBlockFilters = []json.RawMessage{beginFilterEventTypeValid}

	confBytes, err = json.Marshal(conf)
	suite.Require().NoError(err)

	beginBlockFilters, _, _, _, _, err := ParseJSONFilterConfig(confBytes)

	suite.Require().NoError(err)
	suite.Require().Len(beginBlockFilters, 1)
	suite.Require().True(beginBlockFilters[0].EventMatches(filter.EventData{Event: models.BlockEvent{BlockEventType: models.BlockEventType{Type: "coin_received"}}}))
	suite.Require().False(beginBlockFilters[0].EventMatches(filter.EventData{Event: models.BlockEvent{BlockEventType: models.BlockEventType{Type: "dne"}}}))

	conf.BeginBlockFilters = []json.RawMessage{}

	messageTypeFilterInvalid, err := getMockMessageTypeBytes(true)
	suite.Require().NoError(err)

	conf.MessageTypeFilters = []json.RawMessage{messageTypeFilterInvalid}

	confBytes, err = json.Marshal(conf)
	suite.Require().NoError(err)

	_, _, _, _, _, err = ParseJSONFilterConfig(confBytes)
	suite.Require().Error(err)

	messageTypeFilterValid, err := getMockMessageTypeBytes(false)
	suite.Require().NoError(err)

	conf.MessageTypeFilters = []json.RawMessage{messageTypeFilterValid}

	confBytes, err = json.Marshal(conf)
	suite.Require().NoError(err)

	_, _, _, _, messageTypeFilters, err := ParseJSONFilterConfig(confBytes)

	suite.Require().NoError(err)
	suite.Require().Len(messageTypeFilters, 1)
	suite.Require().True(messageTypeFilters[0].MessageTypeMatches(filter.MessageTypeData{MessageType: "/cosmos.bank.v1beta1.MsgSend"}))
	suite.Require().False(messageTypeFilters[0].MessageTypeMatches(filter.MessageTypeData{MessageType: "dne"}))
}

func getMockEventTypeBytes(skipEventTypeKey bool) (json.RawMessage, error) {
	mockEventType := make(map[string]any)

	mockEventType["type"] = "event_type"
	if !skipEventTypeKey {
		mockEventType["event_type"] = "coin_received"
	}

	return json.Marshal(mockEventType)
}

func getMockMessageTypeBytes(skipMessageTypeKey bool) (json.RawMessage, error) {
	mockMessageType := make(map[string]any)

	mockMessageType["type"] = "message_type"
	if !skipMessageTypeKey {
		mockMessageType["message_type"] = "/cosmos.bank.v1beta1.MsgSend"
	}

	return json.Marshal(mockMessageType)
}

func TestFilterConfigTestSuite(t *testing.T) {
	suite.Run(t, new(FilterConfigTestSuite))
}
