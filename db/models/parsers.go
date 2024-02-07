package models

type BlockEventParser struct {
	ID                     uint
	BlockLifecyclePosition BlockLifecyclePosition `gorm:"uniqueIndex:idx_block_event_parser_identifier_lifecycle_position"`
	Identifier             string                 `gorm:"uniqueIndex:idx_block_event_parser_identifier_lifecycle_position"`
}

type BlockEventParserError struct {
	ID                 uint
	BlockEventParserID uint
	BlockEventParser   BlockEventParser
	BlockEventID       uint
	BlockEvent         BlockEvent
	Error              string
}

type MessageParser struct {
	ID uint
	// Should the message type be added here for clarity purposes?
	Identifier string `gorm:"uniqueIndex:idx_message_parser_identifier"`
}

type MessageParserError struct {
	ID              uint
	MessageParserID uint
	MessageParser   MessageParser
	MessageID       uint
	Message         Message
	Error           string
}
