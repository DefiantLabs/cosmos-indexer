#!/bin/sh

# NOTE: I've already run the commands below to produce the Docker images. The only commands you'd need to run in a fresh test environment
# are to import the keys with junod keys add.

# ACCOUNT info:
#
# Validator's delegator address: "juno128taw6wkhfq29u83lmh5qyfv8nff6h0w577vsy"
# Validator address: "junovaloper128taw6wkhfq29u83lmh5qyfv8nff6h0wtrgrta"
#
# MNEMONICS BUILT INTO THIS IMAGE ARE AS FOLLOWS:
# The validator mnemonic is
# wave assume sun shoe wash once unfair master actual vessel diesel actor spend swear elder once fetch spider aim shift brown artefact jump wild
# The kyle test key mnemonic is
# cup lend senior velvet sleep rely stock roast area color violin such urban endless strategy such more future crane cruel tone daring fly style
# Juno development team's built in test key mnemonic is
# clip hire initial neck maid actor venue client foam budget lock catalog sweet steak waste crater broccoli pipe steak sister coyote moment obvious choose

# test key 'kyle' will be used for testing transactions that send/receive funds. use the mnemonic above
junod keys add kyle --recover

# this is the mnemonic of the Juno development team's test user key. it is a genesis account.
junod keys add default --recover

# This is the delegator address that goes with the validator. juno128taw6wkhfq29u83lmh5qyfv8nff6h0w577vsy
junod keys add validator --recover

# Launch the node in the background
docker-compose up -d
# Give the node time to startup in case this is first run
sleep 10

# send some money from the genesis key to our new key (juno1m2hg5t7n8f6kzh8kmh98phenk8a4xp5wyuz34y=the kyle key from above)
junod tx bank send default juno1m2hg5t7n8f6kzh8kmh98phenk8a4xp5wyuz34y 80085ustake --chain-id testing
# show balances
junod query bank balances juno1m2hg5t7n8f6kzh8kmh98phenk8a4xp5wyuz34y --chain-id testing

#Validator address is junovaloper128taw6wkhfq29u83lmh5qyfv8nff6h0wtrgrta in the base image kyle created
junod tx staking delegate junovaloper128taw6wkhfq29u83lmh5qyfv8nff6h0wtrgrta 50000ustake --from kyle --chain-id testing
junod tx staking delegate junovaloper128taw6wkhfq29u83lmh5qyfv8nff6h0wtrgrta 85ustake --from kyle --chain-id testing

#some time later, collect rewards..
junod tx distribution withdraw-rewards junovaloper128taw6wkhfq29u83lmh5qyfv8nff6h0wtrgrta --commission --from validator --chain-id testing

# In the genesis, in the MsgCreateValidator transaction, the delegator starts with 1000000000 and delegates 250000000 leaving 750000000.
# After waiting a while and collecting rewards with the command above, you will always see > 750000000 when you run the query bank balances command.
junod query bank balances juno128taw6wkhfq29u83lmh5qyfv8nff6h0w577vsy --chain-id testing

junod query distribution rewards juno1mt72y3jny20456k247tc5gf2dnat76l4ynvqwl $VALOPER_ADDRESS
junod query distribution commission $VALOPER_ADDRESS
junod query distribution rewards $TEST_USER_ADDRESS $VALOPER_ADDRESS
