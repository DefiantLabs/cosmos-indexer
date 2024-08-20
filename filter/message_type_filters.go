package filter

import (
	"errors"
	"fmt"
	"regexp"
)

type MessageTypeFilter interface {
	MessageTypeMatches(MessageTypeData) (bool, error)
	Ignore() bool
	Valid() (bool, error)
}

type MessageTypeData struct {
	MessageType string
}

type DefaultMessageTypeFilter struct {
	MessageType string `json:"message_type"`
}

type MessageTypeRegexFilter struct {
	MessageTypeRegexPattern string `json:"message_type_regex"`
	messageTypeRegex        *regexp.Regexp
	ShouldIgnore            bool `json:"should_ignore"`
}

func (f DefaultMessageTypeFilter) MessageTypeMatches(messageTypeData MessageTypeData) (bool, error) {
	return messageTypeData.MessageType == f.MessageType, nil
}

func (f MessageTypeRegexFilter) MessageTypeMatches(messageTypeData MessageTypeData) (bool, error) {
	return f.messageTypeRegex.MatchString(messageTypeData.MessageType), nil
}

func (f DefaultMessageTypeFilter) Ignore() bool {
	return false
}

func (f DefaultMessageTypeFilter) Valid() (bool, error) {
	if f.MessageType != "" {
		return true, nil
	}

	return false, errors.New("MessageType must be set")
}

func (f MessageTypeRegexFilter) Valid() (bool, error) {
	if f.messageTypeRegex != nil && f.MessageTypeRegexPattern != "" {
		return true, nil
	}

	return false, errors.New("MessageTypeRegexPattern must be set")
}

func (f MessageTypeRegexFilter) Ignore() bool {
	return f.ShouldIgnore
}

func NewRegexMessageTypeFilter(messageTypeRegexPattern string, shouldIgnore bool) (MessageTypeRegexFilter, error) {
	messageTypeRegex, err := regexp.Compile(messageTypeRegexPattern)
	if err != nil {
		return MessageTypeRegexFilter{}, fmt.Errorf("error compiling message type regex: %s", err)
	}

	return MessageTypeRegexFilter{
		MessageTypeRegexPattern: messageTypeRegexPattern,
		messageTypeRegex:        messageTypeRegex,
		ShouldIgnore:            shouldIgnore,
	}, nil
}
