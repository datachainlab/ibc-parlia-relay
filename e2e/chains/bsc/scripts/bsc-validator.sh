source /scripts/utils.sh

DATA_DIR=/root/.ethereum

wait_for_host_port ${BOOTSTRAP_HOST} ${BOOTSTRAP_TCP_PORT}
BOOTSTRAP_IP=$(get_host_ip $BOOTSTRAP_HOST)
VALIDATOR_ADDR=$(cat ${DATA_DIR}/address)
HOST_IP=$(hostname -i)

echo "validator id: ${HOST_IP}"

ETHSTATS=""
# Use exec to handle signals
exec geth --config ${DATA_DIR}/config.toml --datadir ${DATA_DIR} --netrestrict ${CLUSTER_CIDR} \
	--verbosity ${VERBOSE} --nousb ${ETHSTATS} --state.scheme=hash --db.engine=leveldb \
	--bootnodes enode://${BOOTSTRAP_PUB_KEY}@${BOOTSTRAP_IP}:${BOOTSTRAP_TCP_PORT} \
	--mine --miner.etherbase=${VALIDATOR_ADDR} -unlock ${VALIDATOR_ADDR} --password /dev/null --blspassword /scripts/wallet_password.txt \
	--light.serve 50 --pprof.addr 0.0.0.0 --metrics \
	--rpc.allow-unprotected-txs  --history.transactions 15768000 \
	--pprof --ipcpath /gethipc --vote --override.fixedturnlength 2
