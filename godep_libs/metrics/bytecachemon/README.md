# Metrics for byte cache #

These metrics should be used for caches that realize caching of 
serialized data ([]byte etc.). Implementations may vary but metrics 
are the same.

These metrics are not included to Lazada Metrics Standard yet.
The package declares three metrics for byte cache:

* Hit count.
* Miss count.
* Total number of currently stored items.
