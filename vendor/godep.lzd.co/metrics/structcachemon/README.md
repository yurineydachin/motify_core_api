# Metrics for structure cache #

These metrics should be used for caches that realize caching of structures
contrary to caching of serialized byte arrays. 
There are different realizations of struct caches.

[Lazada Prometheus Metrics Naming Standard (v2.0)](https://confluence.lazada.com/x/OihVAQ)
declares four metrics for struct cache:

* Response time in milliseconds.
* Hit count.
* Miss count.
* Total items currently stored in cache.
