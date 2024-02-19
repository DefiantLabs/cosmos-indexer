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
