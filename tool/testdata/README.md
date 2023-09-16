## Testdata making tool for lcp-parlia

Set bsc rpc addr.

```sh
export BSC_MAINNET_RPC_ADDR="rpc node"
```

### Misbehavior
```sh
# src/client.rs test_success_submit_misbehavior
go run main.go misbehavior success 

# src/client.rs test_error_submit_misbehavior
go run main.go misbehavior error
```

### Header
```sh
# src/client.rs test_success_update_client_epoch
go run main.go update success epoch

# src/client.rs test_success_update_client_non_epoch
go run main.go update success latest

# src/client.rs test_error_update_client
go run main.go update error 
```

### LCP integration data
```sh
go run main.go history mainnet --num 240
go run main.go history testnet --num 240
```