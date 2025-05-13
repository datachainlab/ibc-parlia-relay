#!/usr/bin/env bash

set -e

timestamp=$(geth attach --exec 'parseInt(eth.getBlockByNumber("latest").timestamp)' /gethipc)
[ $(($(date '+%s') - $timestamp)) -lt 5 ]
