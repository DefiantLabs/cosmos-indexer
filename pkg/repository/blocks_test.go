package repository

import (
	"context"
	"fmt"
	"github.com/nodersteam/cosmos-indexer/pkg/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestTotalBlocks(t *testing.T) {
	type expected struct {
		blockHeight int64
		count24H    int64
		totalFees   decimal.Decimal
		err         error
	}

	sampleTxes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1,'random_hash_1', 1003, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2,'random_hash_2', 1003, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3,'random_hash_3', 1003, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4,'random_hash_4', 1001, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (5,'random_hash_5', 1001, 5, '{"signature9", "signature10"}', $1, 'Random memo 5', 500, '{"option9", "option10"}', '{"non_critical_option9", "non_critical_option10"}', 5, 5),
									  (6,'random_hash_6', 1001, 6, '{"signature11", "signature12"}', $1, 'Random memo 6', 600, '{"option11", "option12"}', '{"non_critical_option11", "non_critical_option12"}', 6, 6),
									  (7,'random_hash_7', 1004, 7, '{"signature13", "signature14"}', $1, 'Random memo 7', 700, '{"option13", "option14"}', '{"non_critical_option13", "non_critical_option14"}', 7, 7),
									  (8,'random_hash_8', 1004, 8, '{"signature15", "signature16"}', $1, 'Random memo 8', 800, '{"option15", "option16"}', '{"non_critical_option15", "non_critical_option16"}', 8, 8),
									  (9,'random_hash_9', 1004, 9, '{"signature17", "signature18"}', $1, 'Random memo 9', 900, '{"option17", "option18"}', '{"non_critical_option17", "non_critical_option18"}', 9, 9),
									  (10,'random_hash_10', 1004, 10, '{"signature19", "signature20"}', $1, 'Random memo 10', 1000, '{"option19", "option20"}', '{"non_critical_option19", "non_critical_option20"}', 10, 10);
			`
	sampleBlocks := `INSERT INTO blocks (id, time_stamp, height, chain_id, proposer_cons_address_id, tx_indexed, block_events_indexed, block_hash) 
VALUES 
    (1, $1, 1000, 1, 1, true, true, 'block_hash_1'),
    (2, $1, 1001, 1, 2, true, true, 'block_hash_2'),
    (3, $1, 1002, 1, 3, true, true, 'block_hash_3'),
    (4, $1, 1003, 1, 4, true, true, 'block_hash_4'),
    (5, $1, 1004, 1, 5, true, true, 'block_hash_5');`

	sampleFees := `INSERT INTO fees (tx_id, amount) 
VALUES 
    (5, 10.50),
    (6, 15.75),
    (7, 20.25),
    (8, 12.30),
    (9, 18.90)`

	tests := []struct {
		name   string
		before func()
		to     time.Time
		result expected
		after  func()
	}{
		{"success",
			func() {
				tm := time.Now().Add(-5 * time.Minute).UTC()
				postgresConn.Exec(context.Background(), sampleBlocks, tm)
				postgresConn.Exec(context.Background(), sampleTxes, tm)
				postgresConn.Exec(context.Background(), sampleFees)
			},
			time.Now(),
			expected{blockHeight: 1004, count24H: 5, totalFees: decimal.NewFromInt(11)},
			func() {
				postgresConn.Exec(context.Background(), `delete from blocks`)
				postgresConn.Exec(context.Background(), `delete from txes`)
				postgresConn.Exec(context.Background(), `delete from fees`)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			txsRepo := NewBlocks(postgresConn)
			total, err := txsRepo.TotalBlocks(context.Background(), tt.to)
			require.Equal(t, tt.result.err, err)
			require.Equal(t, tt.result.blockHeight, total.BlockHeight)
			require.Equal(t, tt.result.count24H, total.Count24H)
			require.Equal(t, tt.result.totalFees.String(), total.TotalFee24H.String())
			tt.after()
		})
	}
}

func TestBlocks_GetBlockInfo(t *testing.T) {
	type expected struct {
		bl  model.BlockInfo
		err error
	}

	sampleTxes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1,'random_hash_1', 1003, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2,'random_hash_2', 1003, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3,'random_hash_3', 1003, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4,'random_hash_4', 1001, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (5,'random_hash_5', 1001, 5, '{"signature9", "signature10"}', $1, 'Random memo 5', 500, '{"option9", "option10"}', '{"non_critical_option9", "non_critical_option10"}', 5, 5),
									  (6,'random_hash_6', 1001, 6, '{"signature11", "signature12"}', $1, 'Random memo 6', 600, '{"option11", "option12"}', '{"non_critical_option11", "non_critical_option12"}', 6, 6),
									  (7,'random_hash_7', 1004, 7, '{"signature13", "signature14"}', $1, 'Random memo 7', 700, '{"option13", "option14"}', '{"non_critical_option13", "non_critical_option14"}', 7, 7),
									  (8,'random_hash_8', 1004, 8, '{"signature15", "signature16"}', $1, 'Random memo 8', 800, '{"option15", "option16"}', '{"non_critical_option15", "non_critical_option16"}', 8, 8),
									  (9,'random_hash_9', 1004, 9, '{"signature17", "signature18"}', $1, 'Random memo 9', 900, '{"option17", "option18"}', '{"non_critical_option17", "non_critical_option18"}', 9, 9),
									  (10,'random_hash_10', 1004, 10, '{"signature19", "signature20"}', $1, 'Random memo 10', 1000, '{"option19", "option20"}', '{"non_critical_option19", "non_critical_option20"}', 10, 10);
			`
	sampleBlocks := `INSERT INTO blocks (id, time_stamp, height, chain_id, proposer_cons_address_id, tx_indexed, block_events_indexed, block_hash) 
VALUES 
    (1, $1, 1000, 1, 1, true, true, 'block_hash_1'),
    (2, $1, 1001, 1, 2, true, true, 'block_hash_2'),
    (3, $1, 1002, 1, 3, true, true, 'block_hash_3'),
    (4, $1, 1003, 1, 4, true, true, 'block_hash_4'),
    (5, $1, 1004, 1, 5, true, true, 'block_hash_5');`

	sampleFees := `INSERT INTO fees (tx_id, amount) 
VALUES 
    (5, 110),
    (6, 15.75),
    (7, 20.25),
    (8, 12.30),
    (9, 18.90)`

	tm := time.Now().Add(-5 * time.Minute).UTC()
	_, err := postgresConn.Exec(context.Background(), sampleBlocks, tm)
	require.NoError(t, err)
	_, err = postgresConn.Exec(context.Background(), sampleTxes, tm)
	require.NoError(t, err)
	_, err = postgresConn.Exec(context.Background(), sampleFees)
	require.NoError(t, err)

	defer func() {
		postgresConn.Exec(context.Background(), `delete from blocks`)
		postgresConn.Exec(context.Background(), `delete from txes`)
		postgresConn.Exec(context.Background(), `delete from fees`)
	}()

	tests := []struct {
		name   string
		hash   string
		result expected
	}{
		{"success no fees",
			"block_hash_1",
			expected{bl: model.BlockInfo{BlockHash: "block_hash_1",
				BlockHeight: 1000, TotalTx: 1, TotalFees: decimal.NewFromInt(0)}},
		},
		{"success fees",
			"block_hash_5",
			expected{bl: model.BlockInfo{BlockHash: "block_hash_5",
				BlockHeight: 1004, TotalTx: 1, TotalFees: decimal.RequireFromString("110")}},
		},
		{"not found", "fffff", expected{err: fmt.Errorf("exec no rows in result set")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txsRepo := NewBlocks(postgresConn)
			info, err := txsRepo.GetBlockInfoByHash(context.Background(), tt.hash)
			if tt.result.err != nil && err == nil {
				require.Fail(t, "expected error, got nil")
			}
			require.Equal(t, tt.result.err, err)

			if err == nil {
				require.Equal(t, info.TotalFees.String(), tt.result.bl.TotalFees.String())
				require.Equal(t, info.BlockHash, tt.result.bl.BlockHash)
				require.Equal(t, info.BlockHeight, tt.result.bl.BlockHeight)
				require.Equal(t, info.TotalTx, tt.result.bl.TotalTx)
			}
		})
	}
}
