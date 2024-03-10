package main

import (
	"log"

	"errors"

	"github.com/DefiantLabs/cosmos-indexer/cmd"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/filter"

	"github.com/DefiantLabs/cosmos-indexer/config"
	indexerTxTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/parsers"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
	govV1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govV1Beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"gorm.io/gorm"
)

// This defines the custom message parser for the governance vote message type
// It implements the MessageParser interface
type MsgVoteParser struct {
	Id string
}

func (c *MsgVoteParser) Identifier() string {
	return c.Id
}

func (c *MsgVoteParser) ParseMessage(cosmosMsg stdTypes.Msg, log *indexerTxTypes.LogMessage, cfg config.IndexConfig) (*any, error) {

	msgV1Beta1, okV1Beta1 := cosmosMsg.(*govV1Beta1.MsgVote)
	msgV1, okV1 := cosmosMsg.(*govV1.MsgVote)

	if !okV1Beta1 && !okV1 {
		return nil, errors.New("not a vote message")
	}

	var val Vote

	if okV1Beta1 {
		val = Vote{
			Option: convertV1Beta1VoteOption(msgV1Beta1.Option),
			Address: models.Address{
				Address: msgV1Beta1.Voter,
			},
			Proposal: Proposal{
				ProposalID: msgV1Beta1.ProposalId,
			},
		}
	} else {
		val = Vote{
			Option: convertV1VoteOption(msgV1.Option),
			Address: models.Address{
				Address: msgV1.Voter,
			},
			Proposal: Proposal{
				ProposalID: msgV1.ProposalId,
			},
		}

	}

	storageVal := any(val)

	return &storageVal, nil
}

func convertV1Beta1VoteOption(option govV1Beta1.VoteOption) VoteOption {
	switch option {
	case govV1Beta1.OptionYes:
		return Yes
	case govV1Beta1.OptionNo:
		return No
	case govV1Beta1.OptionAbstain:
		return Abstain
	case govV1Beta1.OptionNoWithVeto:
		return Veto
	case govV1Beta1.OptionEmpty:
		return Empty
	default:
		return -1
	}
}

func convertV1VoteOption(option govV1.VoteOption) VoteOption {
	switch option {
	case govV1.OptionYes:
		return Yes
	case govV1.OptionNo:
		return No
	case govV1.OptionAbstain:
		return Abstain
	case govV1.OptionNoWithVeto:
		return Veto
	case govV1.OptionEmpty:
		return Empty
	default:
		return -1
	}
}

// This method is called during database insertion. It is responsible for storing the parsed data in the database.
// The gorm db is wrapped in a transaction, so any errors will cause a rollback.
// Any errors returned will be saved as a parser error in the database as well for later debugging.
func (c *MsgVoteParser) IndexMessage(dataset *any, db *gorm.DB, message models.Message, messageEvents []parsers.MessageEventWithAttributes, cfg config.IndexConfig) error {
	vote, ok := (*dataset).(Vote)

	if !ok {
		return errors.New("invalid vote type")
	}

	// Find the address in the database
	var err error
	var voter models.Address
	voter, err = dbTypes.FindOrCreateAddressByAddress(db, vote.Address.Address)

	if err != nil {
		return err
	}

	var proposal Proposal
	err = db.Where(&proposal).FirstOrCreate(&proposal).Error

	if err != nil {
		return err
	}

	vote.MsgID = message.ID
	vote.Msg = message
	vote.AddressID = voter.ID
	vote.Address = voter
	vote.ProposalID = proposal.ID
	vote.Proposal = proposal

	err = db.Where(&vote).FirstOrCreate(&vote).Error
	return err
}

// These are the indexer's custom models
// They are used to store the parsed data in the database
type VoteOption int64

const (
	Empty VoteOption = iota
	Yes
	Abstain
	No
	Veto
)

type Vote struct {
	ID         uint
	Msg        models.Message
	MsgID      uint
	AddressID  uint
	Address    models.Address
	Option     VoteOption
	Proposal   Proposal
	ProposalID uint
}

type Proposal struct {
	ID         uint
	ProposalID uint64
}

type GovernanceVotingMessageTypeFilter struct {
}

func main() {

	// Register the custom database models. They will be migrated and included in the database when the indexer finishes setup.
	customModels := []any{}
	customModels = append(customModels, Vote{})
	customModels = append(customModels, Proposal{})

	// This indexer is only concerned with vote messages, so we create a regex filter to only index those messages.
	// This significantly reduces the size of the indexed dataset, saving space and processing time.
	// We use a regex because the message type can be different between v1 and v1beta1 of the gov module.
	govRegexMessageTypeFilter, err := filter.NewRegexMessageTypeFilter("^/cosmos\\.gov.*MsgVote$")

	if err != nil {
		log.Fatalf("Failed to create regex message type filter. Err: %v", err)
	}

	// Register the custom types that will modify the behavior of the indexer
	cmd.RegisterCustomModels(customModels)
	cmd.RegisterMessageTypeFilter(govRegexMessageTypeFilter)

	// Register the custom message parser for the vote message types. Our parser can handle both v1 and v1beta1 vote messages.
	// However, they must be uniquely identified by the Identifier() method. This will make identifying any parser errors easier.
	v1Beta1Parser := &MsgVoteParser{Id: "vote-v1beta1"}
	v1Parser := &MsgVoteParser{Id: "vote-v1"}
	cmd.RegisterCustomMessageParser("/cosmos.gov.v1beta1.MsgVote", v1Beta1Parser)
	cmd.RegisterCustomMessageParser("/cosmos.gov.v1.MsgVote", v1Parser)

	// Execute the root command to start the indexer.
	err = cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
