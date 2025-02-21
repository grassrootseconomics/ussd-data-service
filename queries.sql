--name: last-10-tx
-- Fetches an account's last 10 transfers
-- $1: public_key
SELECT
    token_transfer.sender_address AS sender,
    token_transfer.recipient_address AS recipient,
    token_transfer.transfer_value,
    token_transfer.contract_address,
    tx.tx_hash,
    tx.date_block,
    tx.success,
    tokens.token_symbol,
    tokens.token_decimals
FROM token_transfer
INNER JOIN tx ON token_transfer.tx_id = tx.id
INNER JOIN tokens ON token_transfer.contract_address = tokens.contract_address
WHERE token_transfer.sender_address = $1 OR token_transfer.recipient_address = $1
ORDER BY tx.date_block DESC
LIMIT 10;

--name: token-holdings
-- Fetches an account's token holdings
-- $1: public_key
SELECT DISTINCT tokens.token_symbol, tokens.contract_address, tokens.token_decimals
FROM tokens
LEFT JOIN token_transfer ON tokens.contract_address = token_transfer.contract_address
AND (token_transfer.sender_address = $1 OR token_transfer.recipient_address = $1)
LEFT JOIN token_mint ON tokens.contract_address = token_mint.contract_address
AND (token_mint.minter_address = $1 OR token_mint.recipient_address = $1)
WHERE token_transfer.contract_address IS NOT NULL OR token_mint.contract_address IS NOT NULL;

--name: token-details
-- Fetches token details
-- $1: token_address
SELECT tokens.token_name, tokens.token_symbol, tokens.token_decimals, tokens.sink_address FROM tokens
WHERE tokens.contract_address = $1;

--name: pool-details
-- Fetches tpool details
-- $1: pool_address
SELECT pools.pool_name, pools.pool_symbol, pools.contract_address FROM pools
WHERE pools.contract_address = $1;

--name: pool-reverse-details
-- Fetches pool details
-- $1: pool_symbol
SELECT pools.pool_name, pools.pool_symbol, pools.contract_address FROM pools
WHERE pools.pool_symbol = $1;


--name: top-active-pools
-- Fetches top 5 active pools based on the last 1k swaps
WITH recent_swaps AS (
    SELECT ps.contract_address
    FROM pool_swap ps
    JOIN tx ON ps.tx_id = tx.id
    ORDER BY ps.id DESC
    LIMIT 1000
)
SELECT contract_address, pool_name, pool_symbol
FROM (
    SELECT p.contract_address, p.pool_name, p.pool_symbol, COUNT(*) AS swap_count
    FROM recent_swaps rs
    JOIN pools p ON rs.contract_address = p.contract_address
    WHERE p.removed = FALSE
    GROUP BY p.contract_address, p.pool_name, p.pool_symbol
) sub
ORDER BY sub.swap_count DESC
LIMIT 5;