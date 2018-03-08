# Metrics for Aerospike database #

[Aerospike](http://aerospike.com) is the database used by many Lazada services as a cache
(instead of Memcache) or as a persistent key-value storage for some kinds of data.

[Lazada Prometheus Metrics Naming Standard (v2.0)](https://confluence.lazada.com/x/OihVAQ)
declares four metrics for Aerospike:

* Query duration in milliseconds.
* Hit count (when you using Aerospike as a cache).
* Miss count (when you using Aerospike as a cache).
* Number of connections currently opened to Aerospike.
