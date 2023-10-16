#!/bin/sh
sleep 10

# Import Key

echo $TEST_USER_WALLET_PASS | junod keys import $TEST_USER_WALLET_NAME $TEST_USER_WALLET_KEY
export VALOPER_ADDRESS=$(junod keys show validator -a --bech val)
export TEST_USER_ADDRESS=$(junod keys show $TEST_USER_WALLET_NAME -a)

# Put Test Transactions Here
# Note that the sleep command is needed sometimes after transactions

# NOTE: if you read the Juno team's genesis TXs, you will see that the validator starts with 1000000000 then stakes 250000000 (1/4 of the funds).
# send the test user another 1/4, leaving the validator with 500000000 unstaked.
echo "Y" | junod tx bank send validator $TEST_USER_ADDRESS 250000000ustake --chain-id testing
sleep 2

## Stake almost all of the test user's funds, leaving 1000ustake in their bank
echo "Y" | junod tx staking delegate $VALOPER_ADDRESS 249999000ustake --from $TEST_USER_ADDRESS --chain-id testing
sleep 2

## Only stake 1000ustake of the validator's own funds, leaving the validator with 499999000
echo "Y" | junod tx staking delegate $VALOPER_ADDRESS 1000ustake --from validator --chain-id testing
sleep 2
junod query bank balances $TEST_USER_ADDRESS
