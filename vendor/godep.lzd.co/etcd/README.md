# etcd
A Etcd client creator with Lazada Infra requirements.

- [Usage](#Usage)
- [Installation](#Installation)

## Usage
Just replace constructors from `github.com/coreos/etcd` to constructors from `godep.lzd.co/etcd`:
```go
// Etcd V2
c, err := /*godep.lzd.co*/etcd.NewClient([]string{"https://localhost:2379"})

// Etcd V3
c, err := /*godep.lzd.co*/etcd.NewClientV3([]string{"https://localhost:2379"})
```
`c` will be a client from `github.com/coreos/etcd` with valid settings.

## Installation
`glide get godep.lzd.co/etcd`
