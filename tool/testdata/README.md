## Testdata making tool for lcp-parlia

Set bsc rpc addr.

### Misbehavior
```sh
export BSC_IBC_ADDR=`cat ../../e2e/config/demo/ibc-1.json | jq '.chain.ibc_address'`
export BSC_IBC_ADDR=`echo ${BSC_IBC_ADDR:1:42}`

# src/client.rs test_success_submit_misbehavior
go run main.go misbehavior success 

# src/client.rs test_error_submit_misbehavior
export BSC_RPC_ADDR="rpc node"
go run main.go misbehavior error
```

### Header
```sh
export BSC_RPC_ADDR="http://localhost:8545"
export BSC_IBC_ADDR=`cat ../../e2e/config/demo/ibc-1.json | jq '.chain.ibc_address'`
export BSC_IBC_ADDR=`echo ${BSC_IBC_ADDR:1:42}`

# src/fixture/localnet.rs success_update_client_epoch_input
go run main.go update success epoch

# src/fixture/localnet.rs success_update_client_non_epoch_input
go run main.go update success latest

# src/fixture/localnet.rs success_update_client_continuous_input 
go run main.go update success specified --num 400

# src/fixture/localnet.rs error_update_client_non_neighboring_epoch_input
go run main.go update error 

# src/fixture/localnet.rs headers 
go run main.go header success specified --num 400
```
