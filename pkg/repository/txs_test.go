package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTransactionsPerPeriod(t *testing.T) {
	type expected struct {
		allTx  int64
		all24H int64
		all30D int64
		err    error
	}

	sampleData := `INSERT INTO txes (hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  ('random_hash_1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  ('random_hash_2', 456, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  ('random_hash_3', 789, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  ('random_hash_4', 101112, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  ('random_hash_5', 131415, 5, '{"signature9", "signature10"}', $1, 'Random memo 5', 500, '{"option9", "option10"}', '{"non_critical_option9", "non_critical_option10"}', 5, 5),
									  ('random_hash_6', 161718, 6, '{"signature11", "signature12"}', $1, 'Random memo 6', 600, '{"option11", "option12"}', '{"non_critical_option11", "non_critical_option12"}', 6, 6),
									  ('random_hash_7', 192021, 7, '{"signature13", "signature14"}', $1, 'Random memo 7', 700, '{"option13", "option14"}', '{"non_critical_option13", "non_critical_option14"}', 7, 7),
									  ('random_hash_8', 222324, 8, '{"signature15", "signature16"}', $1, 'Random memo 8', 800, '{"option15", "option16"}', '{"non_critical_option15", "non_critical_option16"}', 8, 8),
									  ('random_hash_9', 252627, 9, '{"signature17", "signature18"}', $1, 'Random memo 9', 900, '{"option17", "option18"}', '{"non_critical_option17", "non_critical_option18"}', 9, 9),
									  ('random_hash_10', 282930, 10, '{"signature19", "signature20"}', $1, 'Random memo 10', 1000, '{"option19", "option20"}', '{"non_critical_option19", "non_critical_option20"}', 10, 10);
			`

	tests := []struct {
		name   string
		before func()
		to     time.Time
		result expected
		after  func()
	}{
		{"success",
			func() {
				postgresConn.Exec(context.Background(), sampleData, time.Now().UTC().Add(-1*time.Hour))
			},
			time.Now().UTC(),
			expected{allTx: 10, all24H: 10, all30D: 10, err: nil},
			func() {
				postgresConn.Exec(context.Background(), `delete from txes`)
			},
		},
		{"success_no24h",
			func() {
				postgresConn.Exec(context.Background(), sampleData, time.Now().UTC().Add(-25*time.Hour))
			},
			time.Now().UTC(),
			expected{allTx: 10, all24H: 0, all30D: 10, err: nil},
			func() {
				postgresConn.Exec(context.Background(), `delete from txes`)
			},
		},
		{"success_no24h_no30d",
			func() {
				postgresConn.Exec(context.Background(), sampleData, time.Now().UTC().Add(-24*31*time.Hour))
			},
			time.Now().UTC(),
			expected{allTx: 10, all24H: 0, all30D: 0, err: nil},
			func() {
				postgresConn.Exec(context.Background(), `delete from txes`)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			txsRepo := NewTxs(postgresConn)
			allTx, all24H, all30D, err := txsRepo.TransactionsPerPeriod(context.Background(), tt.to)
			require.Equal(t, tt.result.err, err)
			require.Equal(t, tt.result.allTx, allTx)
			require.Equal(t, tt.result.all24H, all24H)
			require.Equal(t, tt.result.all30D, all30D)
			tt.after()
		})
	}
}
