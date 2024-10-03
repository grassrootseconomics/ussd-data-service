--name: last-10-tx
-- Fetches an account's last 10 transfers
-- $1: public_key
SELECT
token_transfer.sender_address AS sender, token_transfer.recipient_address AS recipient, token_transfer.transfer_value, token_transfer.contract_address,
tx.tx_hash, tx.date_block,
tokens.token_symbol, tokens.token_decimals
FROM token_transfer
INNER JOIN tx ON token_transfer.tx_id = tx.id
INNER JOIN tokens ON token_transfer.contract_address = tokens.contract_address
WHERE token_transfer.sender_address = $1 OR token_transfer.recipient_address = $1
ORDER BY tx.date_block DESC
LIMIT 10;

--name: token-holdings
-- Fetches an account's token holdings
-- $1: public_key
SELECT DISTINCT tokens.token_symbol, tokens.contract_address, tokens.token_decimals FROM tokens
INNER JOIN token_transfer on tokens.contract_address = token_transfer.contract_address
WHERE token_transfer.sender_address = $1
OR token_transfer.recipient_address = $1;