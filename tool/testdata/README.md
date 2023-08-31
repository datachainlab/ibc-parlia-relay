## Testdata making tool for lcp-parlia

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
go run main.go header success epoch

# src/client.rs test_success_update_client_non_epoch
go run main.go header success latest

# src/client.rs test_error_update_client
go run main.go header error 
```