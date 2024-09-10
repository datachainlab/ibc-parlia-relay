#!/usr/bin/expect
# 10 characters at least wanted
set wallet_password 1234567890

set timeout 5
sleep 10
spawn geth bls account new --datadir [lindex $argv 0]
expect "*assword:*"
send "$wallet_password\r"
expect "*assword:*"
send "$wallet_password\r"
expect EOF