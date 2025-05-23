#!/usr/bin/env bash

source /scripts/utils.sh

DATA_DIR=/root/.ethereum

account_cnt=$(ls ${DATA_DIR}/keystore | wc -l)
i=1
unlock_sequences="0"
while [ "$i" -lt ${account_cnt} ]; do
	unlock_sequences="${unlock_sequences},${i}"
	i=$((i + 1))
done

ETHSTATS=""
# Use exec to handle signals
exec geth --config ${DATA_DIR}/config.toml --datadir ${DATA_DIR} --netrestrict ${CLUSTER_CIDR} \
	--state.scheme=hash --db.engine=leveldb --verbosity ${VERBOSE} --nousb ${ETHSTATS} \
	--unlock ${unlock_sequences} --password /dev/null --ipcpath /gethipc --override.fixedturnlength 2
