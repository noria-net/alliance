#!/bin/bash

# This script wipe your config folder (~/.noria),
# creates a new wallet named "me"
# and prepares everything to be able to start running 
# a fresh chain from height 1.
# 
# This is not meant to be used when trying to sync to an existing chain,
# but rather to work in a local development environment.

BINARY_HOME="$HOME/.alliance"
CONFIG_HOME="$BINARY_HOME/config"
CHAIN_ID="alliance-1"
DENOM="stake"
GAS_PRICE="0.0025"

echo -e "\nRemoving previous config folder ($BINARY_HOME)"
rm -rf $BINARY_HOME

# Set your keyring (the thing that saves your private keys) to the ~/.noria folder (not secure, only use for testing env)
echo "Setting keyring to \"test\""
allianced config keyring-backend test

# Set the default chain to use
echo "Setting chain-id to \"$CHAIN_ID\""
allianced config chain-id $CHAIN_ID

# Create a new wallet named "me"
allianced keys add me
ADDR=$(allianced keys show me -a)

# Initialize a new genesis.json file
allianced init me --overwrite --chain-id $CHAIN_ID

# Add your freshly created account to the new chain genesis
allianced add-genesis-account me 1000000000$DENOM

# Generate the genesis transaction to create a new validator
allianced gentx me 100000000$DENOM --chain-id $CHAIN_ID --commission-rate 0.1 --commission-max-rate 0.2 --commission-max-change-rate 0.01

# Add that gentx transaction to the genesis file
allianced collect-gentxs

# Edit genesis
sed -i.bak "s/stake/$DENOM/g" $CONFIG_HOME/genesis.json
sed -i.bak 's/"inflation": "[^"]*"/"inflation": "0\.0"/g' $CONFIG_HOME/genesis.json
sed -i.bak 's/"inflation_rate_change": "[^"]*"/"inflation_rate_change": "0\.0"/g' $CONFIG_HOME/genesis.json
sed -i.bak 's/"inflation_min": "[^"]*"/"inflation_min": "0\.0"/g' $CONFIG_HOME/genesis.json
sed -i.bak 's/"voting_period": "[^"]*"/"voting_period": "5s"/g' $CONFIG_HOME/genesis.json
sed -i.bak 's/"quorum": "[^"]*"/"quorum": "0.000001"/g' $CONFIG_HOME/genesis.json
sed -i.bak 's/"reward_delay_time": "[^"]*"/"reward_delay_time": "1s"/g' $CONFIG_HOME/genesis.json

rm $CONFIG_HOME/genesis.json.bak

# Edit config.toml to set the block speed to 1s
sed -i.bak 's/^timeout_commit\ =\ .*/timeout_commit\ =\ \"1s\"/g' $CONFIG_HOME/config.toml
rm $CONFIG_HOME/config.toml.bak

# Edit app.toml to set the minimum gas price
sed -i.bak "s/^minimum-gas-prices\ =\ .*/minimum-gas-prices\ =\ \"0.0025$DENOM\"/g" $CONFIG_HOME/app.toml

# Edit app.toml to enable LCD REST server on port 1317 and REST documentation at http://localhost:1317/swagger/
sed -i.bak 's/^enable\ =\ false/enable\ =\ true/g' $CONFIG_HOME/app.toml
sed -i.bak 's/^swagger\ =\ false/swagger\ =\ true/g' $CONFIG_HOME/app.toml
sed -i.bak '/Rosetta API/{n; s/true/false/}' $CONFIG_HOME/app.toml
sed -i.bak "s/^enabled-unsafe-cors\ =\ .*/enabled-unsafe-cors\ =\ true/g" $CONFIG_HOME/app.toml
sed -i.bak "s/localhost/0.0.0.0/g" $CONFIG_HOME/app.toml
sed -i.bak "s/127\.0\.0\.1/0.0.0.0/g" $CONFIG_HOME/config.toml
rm $CONFIG_HOME/app.toml.bak

echo -e "\n\nYou can now start your chain with 'allianced start'\n"
