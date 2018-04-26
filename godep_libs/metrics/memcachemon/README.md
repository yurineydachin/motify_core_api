# Metrics for memcache #

Memcache (https://memcached.org/) is key-value storage.

[Lazada Prometheus Metrics Naming Standard (v2.0)](https://confluence.lazada.com/x/OihVAQ)
declares three metrics for struct cache:

* Response time in milliseconds.
* Hit count (when you using Aerospike as a cache).
* Miss count (when you using Aerospike as a cache).
