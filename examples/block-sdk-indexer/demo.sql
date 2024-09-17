SELECT b.height, tx.hash, met.message_type from txes tx
    JOIN messages me on me.tx_id = tx.id
    JOIN blocks b on b.id = tx.block_id
    JOIN message_types met on met.id = me.message_type_id;