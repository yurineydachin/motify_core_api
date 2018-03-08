## Registration and balancing example

Example project, showing how to use registrators and balancers.
Runs 2 services with 2 instances each, locating each other through etcd3 provider.

### Installation and running

Use Makefile, Luke!

It's assumed that Golang 1.7+ is pre-installed.

Also Docker should be installed to run Etcd.
Otherwise you can specify existing etcd3 endpoints to the binary

And you may need to install library glide dependencies via `glide install` from the
repository root.

- Run etcd instance in docker

    `make etcd-run-cluster`

- Build binary

    `make build`

- Run services

    `make run` runs 2 services with 3 instances each

    `make kill` kills started services

- Or all together in single command!

    `make run-build`
