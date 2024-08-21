select b.height, a.address, p.proposal_id, CASE 
        WHEN vo.option = 1 THEN 'yes' 
        WHEN vo.option = 2 THEN 'abstain'
        WHEN vo.option = 3 THEN 'no'
        WHEN vo.option = 4 THEN 'veto'
        ELSE 'empty' END as vote
    FROM votes vo
    JOIN messages me on me.id = vo.msg_id
    JOIN proposals p on p.id = vo.proposal_id
    JOIN txes tx on tx.id = me.tx_id
    JOIN blocks b on b.id = tx.block_id
    JOIN message_types met on met.id = me.message_type_id
    JOIN addresses a on vo.address_id = a.id;