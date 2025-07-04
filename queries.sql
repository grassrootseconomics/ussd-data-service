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
FROM chain_data.token_transfer
INNER JOIN chain_data.tx ON token_transfer.tx_id = tx.id
INNER JOIN chain_data.tokens ON token_transfer.contract_address = tokens.contract_address
WHERE token_transfer.sender_address = $1 OR token_transfer.recipient_address = $1
ORDER BY tx.date_block DESC
LIMIT 10;

--name: token-holdings
-- Fetches an account's token holdings
-- $1: public_key
SELECT DISTINCT tokens.token_symbol, tokens.contract_address, tokens.token_decimals
FROM chain_data.tokens
LEFT JOIN chain_data.token_transfer ON tokens.contract_address = token_transfer.contract_address
AND (token_transfer.sender_address = $1 OR token_transfer.recipient_address = $1)
LEFT JOIN chain_data.token_mint ON tokens.contract_address = token_mint.contract_address
AND (token_mint.minter_address = $1 OR token_mint.recipient_address = $1)
WHERE token_transfer.contract_address IS NOT NULL OR token_mint.contract_address IS NOT NULL;

--name: token-details
-- Fetches token details
-- $1: token_address
SELECT tokens.contract_address AS token_address, tokens.token_name, tokens.token_symbol, tokens.token_decimals, tokens.sink_address FROM chain_data.tokens
WHERE tokens.contract_address = $1;

--name: pool-details
-- Fetches pool details from pool_router schema
-- $1: pool_address
SELECT 
    pool_name, 
    pool_symbol, 
    pool_address as contract_address,
    token_registry_address,
    token_limiter_address
FROM pool_router.swap_pools 
WHERE pool_address = $1;

--name: pool-reverse-details
-- Fetches pool details by symbol from pool_router schema
-- $1: pool_symbol
SELECT 
    pool_name, 
    pool_symbol, 
    pool_address as contract_address,
    token_registry_address,
    token_limiter_address
FROM pool_router.swap_pools 
WHERE pool_symbol = $1;


--name: top-active-pools
-- Fetches top 5 active pools based on the last 1k swaps
WITH recent_swaps AS (
    SELECT ps.contract_address
    FROM chain_data.pool_swap ps
    JOIN chain_data.tx ON ps.tx_id = tx.id
    ORDER BY ps.id DESC
    LIMIT 1000
)
SELECT 
    contract_address, 
    pool_name, 
    pool_symbol,
    token_registry_address,
    token_limiter_address
FROM (
    SELECT 
        p.pool_address as contract_address, 
        p.pool_name, 
        p.pool_symbol, 
        p.token_registry_address,
        p.token_limiter_address,
        COUNT(*) AS swap_count
    FROM recent_swaps rs
    JOIN pool_router.swap_pools p ON rs.contract_address = p.pool_address
    GROUP BY p.pool_address, p.pool_name, p.pool_symbol, p.token_registry_address, p.token_limiter_address
) sub
ORDER BY sub.swap_count DESC
LIMIT 5;

--name: pool-token-allowed
-- Checks if a token is allowed in a specific pool
-- $1: pool_address
-- $2: token_address
SELECT EXISTS (
    SELECT 1 
    FROM pool_router.pool_allowed_tokens 
    WHERE pool_address = $1 AND token_address = $2
) AS is_allowed;

--name: pool-allowed-tokens-for-user
-- Fetches user's token holdings that are allowed in a specific pool
-- $1: user_address
-- $2: pool_address
SELECT DISTINCT 
    t.token_symbol, 
    t.contract_address, 
    t.token_decimals
FROM chain_data.tokens t
INNER JOIN pool_router.pool_allowed_tokens pat ON t.contract_address = pat.token_address
LEFT JOIN chain_data.token_transfer tt ON t.contract_address = tt.contract_address
    AND (tt.sender_address = $1 OR tt.recipient_address = $1)
LEFT JOIN chain_data.token_mint tm ON t.contract_address = tm.contract_address
    AND (tm.minter_address = $1 OR tm.recipient_address = $1)
WHERE pat.pool_address = $2
    AND (tt.contract_address IS NOT NULL OR tm.contract_address IS NOT NULL);

--name: pool-allowed-tokens
-- Fetches all tokens allowed in a specific pool
-- $1: pool_address
SELECT DISTINCT 
    t.token_symbol, 
    t.token_address as contract_address, 
    t.token_decimals
FROM pool_router.tokens t
INNER JOIN pool_router.pool_allowed_tokens pat ON t.token_address = pat.token_address
WHERE pat.pool_address = $1;

--name: pool-allowed-stables
-- Fetches stable tokens allowed in a specific pool
-- $1: pool_address
SELECT DISTINCT 
    t.token_symbol, 
    t.token_address as contract_address, 
    t.token_decimals
FROM pool_router.tokens t
INNER JOIN pool_router.pool_allowed_tokens pat ON t.token_address = pat.token_address
WHERE pat.pool_address = $1 
    AND (t.token_symbol = 'cUSD' OR t.token_symbol = 'cKES');

--name: pool-token-swap-rates
-- Fetches exchange rates, decimals, and token limit for two tokens in a specific pool
-- $1: pool_address
-- $2: in_token_address
-- $3: out_token_address
SELECT
    in_token.exchange_rate as in_rate,
    out_token.exchange_rate as out_rate,
    in_token_details.token_decimals as in_decimals,
    out_token_details.token_decimals as out_decimals,
    COALESCE(in_token_limit.token_limit, '0') as in_token_limit,
    COALESCE(out_token_limit.token_limit, '0') as out_token_limit
FROM pool_router.pool_token_exchange_rates in_token
JOIN pool_router.pool_token_exchange_rates out_token
    ON in_token.pool_address = out_token.pool_address
JOIN pool_router.tokens in_token_details
    ON in_token.token_address = in_token_details.token_address
JOIN pool_router.tokens out_token_details
    ON out_token.token_address = out_token_details.token_address
LEFT JOIN pool_router.pool_token_limits in_token_limit
    ON in_token.pool_address = in_token_limit.pool_address
    AND in_token.token_address = in_token_limit.token_address
LEFT JOIN pool_router.pool_token_limits out_token_limit
    ON out_token.pool_address = out_token_limit.pool_address
    AND out_token.token_address = out_token_limit.token_address
WHERE in_token.pool_address = $1
    AND in_token.token_address = $2
    AND out_token.token_address = $3;