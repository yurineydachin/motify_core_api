# etcd
A Etcd client creator with Lazada Infra requirements.

- [Usage](#Usage)
- [Installation](#Installation)

## Usage
Just replace constructors from `github.com/coreos/etcd` to constructors from `motify_core_api/godep_libs/etcd`:
```go
// Etcd V2
c, err := /*motify_core_api/godep_libs*/etcd.NewClient([]string{"https://localhost:2379"})

// Etcd V3
c, err := /*motify_core_api/godep_libs*/etcd.NewClientV3([]string{"https://localhost:2379"})
```
`c` will be a client from `github.com/coreos/etcd` with valid settings.

## Installation
`glide get motify_core_api/godep_libs/etcd`
