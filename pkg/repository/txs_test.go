package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
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
			allTx, all24H, _, all30D, err := txsRepo.TransactionsPerPeriod(context.Background(), tt.to)
			require.Equal(t, tt.result.err, err)
			require.Equal(t, tt.result.allTx, allTx)
			require.Equal(t, tt.result.all24H, all24H)
			require.Equal(t, tt.result.all30D, all30D)
			tt.after()
		})
	}
}

func TestTxs_TransactionRawLog(t *testing.T) {
	ctx := context.Background()

	type expected struct {
		rawLog string
		err    error
	}

	type params struct {
		txHash string
	}

	txResponses := `
					INSERT INTO tx_responses (id, tx_hash, height, time_stamp, code, raw_log, gas_used, gas_wanted, codespace, data, info)
					VALUES 
					  (1, 'hash1', '1234', '2024-04-22 12:00:00', 0, 'raw_log_1', 100, 200, 'codespace1', 'data1', 'info1'),
					  (2, 'hash2', '1235', '2024-04-23 12:00:00', 1, 'raw_log_2', 150, 250, 'codespace2', 'data2', 'info2'),
					  (3, 'hash3', '1236', '2024-04-24 12:00:00', 2, 'raw_log_3', 200, 300, 'codespace3', 'data3', 'info3');
					`
	txes := `INSERT INTO txes (hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  ('hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  ('hash2', 456, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  ('hash3', 789, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  ('hash4', 101112, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	tests := []struct {
		name     string
		expected expected
		params   params
		before   func()
		after    func()
	}{
		{
			"success",
			expected{rawLog: "raw_log_1", err: nil},
			params{"hash1"},
			func() {
				postgresConn.Exec(ctx, txResponses)
				postgresConn.Exec(ctx, txes, time.Now().UTC())
			},
			func() {
				postgresConn.Exec(context.Background(), `delete from txes`)
				postgresConn.Exec(context.Background(), `delete from tx_responses`)
			},
		},
		{
			"not_found",
			expected{err: fmt.Errorf("not found")},
			params{"hash7"},
			func() {
				postgresConn.Exec(ctx, txResponses)
				postgresConn.Exec(ctx, txes, time.Now().UTC())
			},
			func() {
				postgresConn.Exec(context.Background(), `delete from txes`)
				postgresConn.Exec(context.Background(), `delete from tx_responses`)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before()
			txsRepo := NewTxs(postgresConn)
			res, err := txsRepo.TransactionRawLog(context.Background(), tt.params.txHash)
			require.Equal(t, tt.expected.err, err)
			if err == nil {
				require.Equal(t, tt.expected.rawLog, string(res))
			}
			tt.after()
		})
	}
}

func TestTxs_TransactionSigners(t *testing.T) {
	authInfoDemo := `INSERT INTO tx_auth_info (id, fee_id, tip_id) values (1, 1, 1)`
	_, err := postgresConn.Exec(context.Background(), authInfoDemo)
	require.NoError(t, err)

	signerInfos := `INSERT INTO tx_signer_infos (auth_info_id, signer_info_id) values (1, 2)`
	_, err = postgresConn.Exec(context.Background(), signerInfos)
	require.NoError(t, err)

	signerInfo := `INSERT INTO tx_signer_info (id, address_id) values (2, 4)`
	_, err = postgresConn.Exec(context.Background(), signerInfo)
	require.NoError(t, err)

	addresses := `INSERT INTO addresses (id, address) values (4, 'test')`
	_, err = postgresConn.Exec(context.Background(), addresses)
	require.NoError(t, err)

	txes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1, 'hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2, 'hash2', 456, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3, 'hash3', 789, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4, 'hash4', 101112, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	_, err = postgresConn.Exec(context.Background(), txes, time.Now().UTC())
	require.NoError(t, err)

	defer func() {
		postgresConn.Exec(context.Background(), `delete from txes`)
		postgresConn.Exec(context.Background(), `delete from addresses`)
		postgresConn.Exec(context.Background(), `delete from tx_signer_addresses`)
		postgresConn.Exec(context.Background(), `delete from tx_signer_info`)
		postgresConn.Exec(context.Background(), `delete from tx_signer_infos`)
		postgresConn.Exec(context.Background(), `delete from tx_auth_info`)
	}()

	txsRepo := NewTxs(postgresConn)
	res, err := txsRepo.TransactionSigners(context.Background(), "hash1")
	require.NoError(t, err)
	require.Len(t, res, 1)
	require.NotNil(t, res[0].Address)
	require.Equal(t, res[0].Address.Address, "test")
}

func TestTxs_Transactions_ByHash(t *testing.T) {
	defer func() {
		postgresConn.Exec(context.Background(), `delete from txes`)
	}()

	txes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1, 'hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2, 'hash2', 456, 2, '{"signature3", "signature4"}', $1, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3, 'hash3', 789, 3, '{"signature5", "signature6"}', $1, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4, 'hash4', 101112, 4, '{"signature7", "signature8"}', $1, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	_, err := postgresConn.Exec(context.Background(), txes, time.Now().UTC())
	require.NoError(t, err)
	txsRepo := NewTxs(postgresConn)
	txHash := "hash1"

	res, _, err := txsRepo.Transactions(context.Background(), 100, 0, &TxsFilter{TxHash: &txHash})
	require.NoError(t, err)
	require.Len(t, res, 1)
}

func TestTxs_ChartTransactionsByHour(t *testing.T) {
	defer func() {
		postgresConn.Exec(context.Background(), `delete from txes`)
	}()

	txes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1, 'hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2, 'hash2', 456, 2, '{"signature3", "signature4"}', $2, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3, 'hash3', 789, 3, '{"signature5", "signature6"}', $3, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4, 'hash4', 101112, 4, '{"signature7", "signature8"}', $4, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (5, 'hash5', 101112, 5, '{"signature7", "signature8"}', $4, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (6, 'hash6', 101112, 5, '{"signature7", "signature8"}', $5, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	initTime := time.Now().UTC()
	_, err := postgresConn.Exec(context.Background(), txes,
		initTime,
		initTime.Add(-1*time.Hour),
		initTime.Add(-2*time.Hour),
		initTime.Add(-3*time.Hour),
		initTime.Add(-35*time.Hour))
	require.NoError(t, err)
	txsRepo := NewTxs(postgresConn)
	res, err := txsRepo.ChartTransactionsByHour(context.Background(), initTime.Add(5*time.Minute))
	require.NoError(t, err)
	require.NotNil(t, res)

	require.Equal(t, res.Total24H, int64(5))
	require.Equal(t, res.Total48H, int64(1))
	require.Len(t, res.Points, 4)
}

func TestTxs_ChartTransactionsVolume(t *testing.T) {
	defer func() {
		postgresConn.Exec(context.Background(), `delete from txes`)
		postgresConn.Exec(context.Background(), `delete from fees`)
		postgresConn.Exec(context.Background(), `delete from denoms`)
	}()

	batch := pgx.Batch{}

	txes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1, 'hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2, 'hash2', 456, 2, '{"signature3", "signature4"}', $2, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3, 'hash3', 789, 3, '{"signature5", "signature6"}', $3, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4, 'hash4', 101112, 4, '{"signature7", "signature8"}', $4, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (5, 'hash5', 101112, 5, '{"signature7", "signature8"}', $4, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (6, 'hash6', 101112, 5, '{"signature7", "signature8"}', $5, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	initTime := time.Now().UTC()
	batch.Queue(txes, initTime,
		initTime.Add(-1*time.Hour),
		initTime.Add(-2*time.Hour),
		initTime.Add(-3*time.Hour),
		initTime.Add(-35*time.Hour))

	denoms := `INSERT INTO denoms(id, base) VALUES (1, 'utia')`
	batch.Queue(denoms)

	fees := `INSERT INTO fees(id, tx_id, amount, denomination_id)
							VALUES 
								(1, 1, 1000, 1),
								(2, 2, 2000, 1),
								(3, 3, 100, 1),
								(4, 4, 2000, 1),
								(5, 5, 9, 1),
								(6, 6, 27, 1)
							`
	batch.Queue(fees)
	res := postgresConn.SendBatch(context.Background(), &batch)
	defer func(res pgx.BatchResults) {
		err := res.Close()
		require.NoError(t, err)
	}(res)
	for i := 0; i < batch.Len(); i++ {
		_, err := res.Exec()
		if err != nil {
			require.NoError(t, err)
		}
	}

	txsRepo := NewTxs(postgresConn)
	data, err := txsRepo.ChartTransactionsVolume(context.Background(), initTime.Add(5*time.Minute))
	require.NoError(t, err)
	require.Len(t, data, 4)

	require.Equal(t, data[0].TxVolume, decimal.RequireFromString("2009"))
	require.Equal(t, data[1].TxVolume, decimal.RequireFromString("100"))
	require.Equal(t, data[2].TxVolume, decimal.RequireFromString("2000"))
	require.Equal(t, data[3].TxVolume, decimal.RequireFromString("1000"))
}

func TestTxs_VolumePerPeriod(t *testing.T) {
	defer func() {
		postgresConn.Exec(context.Background(), `delete from txes`)
		postgresConn.Exec(context.Background(), `delete from fees`)
		postgresConn.Exec(context.Background(), `delete from denoms`)
	}()

	batch := pgx.Batch{}

	txes := `INSERT INTO txes (id, hash, code, block_id, signatures, timestamp, memo, timeout_height, extension_options, non_critical_extension_options, auth_info_id, tx_response_id)
									VALUES
									  (1, 'hash1', 123, 1, '{"signature1", "signature2"}', $1, 'Random memo 1', 100, '{"option1", "option2"}', '{"non_critical_option1", "non_critical_option2"}', 1, 1),
									  (2, 'hash2', 456, 2, '{"signature3", "signature4"}', $2, 'Random memo 2', 200, '{"option3", "option4"}', '{"non_critical_option3", "non_critical_option4"}', 2, 2),
									  (3, 'hash3', 789, 3, '{"signature5", "signature6"}', $3, 'Random memo 3', 300, '{"option5", "option6"}', '{"non_critical_option5", "non_critical_option6"}', 3, 3),
									  (4, 'hash4', 101112, 4, '{"signature7", "signature8"}', $4, 'Random memo 4', 400, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (5, 'hash5', 101112, 5, '{"signature7", "signature8"}', $4, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4),
									  (6, 'hash6', 101112, 5, '{"signature7", "signature8"}', $5, 'Random memo 5', 600, '{"option7", "option8"}', '{"non_critical_option7", "non_critical_option8"}', 4, 4)
									  `
	initTime := time.Now().UTC()
	batch.Queue(txes, initTime,
		initTime.Add(-1*time.Hour),
		initTime.Add(-2*time.Hour),
		initTime.Add(-3*time.Hour),
		initTime.Add(-35*time.Hour))

	denoms := `INSERT INTO denoms(id, base) VALUES (1, 'utia')`
	batch.Queue(denoms)

	fees := `INSERT INTO fees(id, tx_id, amount, denomination_id)
							VALUES 
								(1, 1, 1000, 1),
								(2, 2, 2000, 1),
								(3, 3, 100, 1),
								(4, 4, 2000, 1),
								(5, 5, 9, 1),
								(6, 6, 27, 1)
							`
	batch.Queue(fees)
	res := postgresConn.SendBatch(context.Background(), &batch)
	defer func(res pgx.BatchResults) {
		err := res.Close()
		require.NoError(t, err)
	}(res)
	for i := 0; i < batch.Len(); i++ {
		_, err := res.Exec()
		if err != nil {
			require.NoError(t, err)
		}
	}

	txsRepo := NewTxs(postgresConn)
	total24H, total30D, err := txsRepo.VolumePerPeriod(context.Background(),
		initTime.Add(5*time.Minute))
	require.NoError(t, err)
	require.Equal(t, total24H, decimal.RequireFromString("5109"))
	require.Equal(t, total30D, decimal.RequireFromString("5136"))
}
