# Example Chaos Experiments

## Quickstart
You can run these examples either of the following ways. Before setting up your environment be sure you have setup a kubernetes backend with chaos mesh.

Also ensure that you have no running enclaves. (`kurtosis enclave ls` should be empty. You can force remove listed enclaves via: `kurtosis enclave rm -f <enclave_identifier>`)

Make sure you have run `kurtosis gateway` in another terminal

### go test
1. Stop all running enclaves
2. Start a new kurtosis enclave with the example config in this dir.
```bash
kurtosis run --enclave example-chaos-experiment github.com/kurtosis-tech/ethereum-package "$(cat example-kurtosis-config.yaml)"
```


Once the enclave is running you can run the tests in `service_test.go` for attacknet.


### attacknet binary
