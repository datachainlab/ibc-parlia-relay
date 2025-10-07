#!/bin/bash
#set wallet_password 1234567890

echo "$1"
geth bls account new --datadir "$1"
sleep 2