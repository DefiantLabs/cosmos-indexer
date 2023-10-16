package tx

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const EventAttributeAmount = "amount"

func GetMessageLogForIndex(logs []LogMessage, index int) *LogMessage {
	for _, log := range logs {
		if log.MessageIndex == index {
			return &log
		}
	}

	return nil
}

func GetEventWithType(eventType string, msg *LogMessage) *LogMessageEvent {
	if msg == nil || msg.Events == nil {
		return nil
	}

	for _, logEvent := range msg.Events {
		if logEvent.Type == eventType {
			return &logEvent
		}
	}

	return nil
}

func GetAllEventsWithType(eventType string, msg *LogMessage) []LogMessageEvent {
	logEventMessages := []LogMessageEvent{}

	if msg == nil || msg.Events == nil {
		return logEventMessages
	}

	for _, logEvent := range msg.Events {
		if logEvent.Type == eventType {
			logEventMessages = append(logEventMessages, logEvent)
		}
	}

	return logEventMessages
}

func GetEventsWithType(eventType string, msg *LogMessage) []LogMessageEvent {
	events := []LogMessageEvent{}
	if msg == nil || msg.Events == nil {
		return nil
	}

	for _, logEvent := range msg.Events {
		if logEvent.Type == eventType {
			events = append(events, logEvent)
		}
	}

	return events
}

type TransferEvent struct {
	Recipient string
	Sender    string
	Amount    string
}

// Transfer events should have attributes in the order recipient, sender, amount.
func ParseTransferEvent(evt LogMessageEvent) ([]TransferEvent, error) {
	errInvalidTransfer := errors.New("not a valid transfer event")
	transfers := []TransferEvent{}
	if evt.Type != "transfer" {
		return nil, errInvalidTransfer
	}

	for i := 0; i < len(evt.Attributes); i++ {
		attrRecipient := evt.Attributes[i]
		if attrRecipient.Key == "recipient" {
			attrSenderIdx := i + 1
			attrAmountIdx := i + 2
			if attrAmountIdx < len(evt.Attributes) {
				attrSender := evt.Attributes[attrSenderIdx]
				attrAmount := evt.Attributes[attrAmountIdx]
				if attrSender.Key == "sender" && attrAmount.Key == EventAttributeAmount {
					transfers = append(transfers, TransferEvent{
						Recipient: attrRecipient.Value,
						Sender:    attrSender.Value,
						Amount:    attrAmount.Value,
					})
				} else {
					return nil, errInvalidTransfer
				}
			} else {
				return nil, errInvalidTransfer
			}
		} else if i%3 == 0 { // every third attr should be "recipient"
			return nil, errInvalidTransfer
		}
	}

	return transfers, nil
}

// If order is reversed, the last attribute containing the given key will be returned
// otherwise the first attribute will be returned
func GetValueForAttribute(key string, evt *LogMessageEvent) (string, error) {
	if evt == nil || evt.Attributes == nil {
		return "", nil
	}

	for _, attr := range evt.Attributes {
		if attr.Key == key {
			return attr.Value, nil
		}
	}

	return "", fmt.Errorf("Attribute %s missing from event", key)
}

func GetCoinsSpent(spender string, evts []LogMessageEvent) []string {
	coinsSpent := []string{}

	if len(evts) == 0 {
		return coinsSpent
	}

	for _, evt := range evts {
		for i := 0; i < len(evt.Attributes); i++ {
			attr := evt.Attributes[i]
			if attr.Key == "spender" && attr.Value == spender {
				attrAmountIdx := i + 1
				if attrAmountIdx < len(evt.Attributes) {
					attrNext := evt.Attributes[attrAmountIdx]
					if attrNext.Key == EventAttributeAmount {
						commaSeperatedCoins := attrNext.Value
						currentCoins := strings.Split(commaSeperatedCoins, ",")
						for _, coin := range currentCoins {
							if coin != "" {
								coinsSpent = append(coinsSpent, coin)
							}
						}
					}
				}
			}
		}
	}

	return coinsSpent
}

func GetCoinsReceived(receiver string, evts []LogMessageEvent) []string {
	coinsReceived := []string{}

	if len(evts) == 0 {
		return coinsReceived
	}

	for _, evt := range evts {
		for i := 0; i < len(evt.Attributes); i++ {
			attr := evt.Attributes[i]
			if attr.Key == "receiver" && attr.Value == receiver {
				attrAmountIdx := i + 1
				if attrAmountIdx < len(evt.Attributes) {
					attrNext := evt.Attributes[attrAmountIdx]
					if attrNext.Key == EventAttributeAmount {
						commaSeperatedCoins := attrNext.Value
						currentCoins := strings.Split(commaSeperatedCoins, ",")
						for _, coin := range currentCoins {
							if coin != "" {
								coinsReceived = append(coinsReceived, coin)
							}
						}
					}
				}
			}
		}
	}

	return coinsReceived
}

// Get the Nth value for the given key (starting at 1)
func GetNthValueForAttribute(key string, n int, evt *LogMessageEvent) string {
	if evt == nil || evt.Attributes == nil {
		return ""
	}
	var count int
	for i := 0; i < len(evt.Attributes); i++ {
		attr := evt.Attributes[i]
		if attr.Key == key {
			count++
			if count == n {
				return attr.Value
			}
		}
	}

	return ""
}

func GetLastValueForAttribute(key string, evt *LogMessageEvent) string {
	if evt == nil || evt.Attributes == nil {
		return ""
	}

	for i := len(evt.Attributes) - 1; i >= 0; i-- {
		attr := evt.Attributes[i]
		if attr.Key == key {
			return attr.Value
		}
	}

	return ""
}

func IsMessageActionEquals(msgType string, msg *LogMessage) bool {
	logEvent := GetEventWithType("message", msg)
	altMsgType := getAltMsgType(msgType)
	if logEvent == nil {
		return false
	}

	for _, attr := range logEvent.Attributes {
		if attr.Key == "action" {
			if attr.Value == msgType || attr.Value == altMsgType {
				return true
			}
		}
	}

	return false
}

var altMsgMap = map[string]string{
	"/cosmos.staking.v1beta1.MsgUndelegate": "begin_unbonding",
}

func getAltMsgType(msgType string) string {
	if altMsg, ok := altMsgMap[msgType]; ok {
		return altMsg
	}

	var output string
	msgParts := strings.Split(msgType, ".Msg")
	if len(msgParts) == 2 {
		msgSuffix := msgParts[1]
		for i, char := range msgSuffix {
			if unicode.IsUpper(char) {
				if i != 0 {
					output = fmt.Sprintf("%v_", output)
				}
			}
			output = fmt.Sprintf("%v%v", output, string(unicode.ToLower(char)))
		}
	}
	return output
}
