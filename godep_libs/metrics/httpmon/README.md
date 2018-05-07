# Handlers (HTTP) metrics for Prometheus #

Application metrics for HTTP handlers. They applicable for non-HTTP handlers too.
So if the service uses something completely different, for example protocol buffers, 
it anyway should use these metrics for measuring response time etc.

[Lazada Prometheus Metrics Naming Standard (v2.0)](https://confluence.lazada.com/x/OihVAQ)
declares two metrics for handlers

* Response time in milliseconds (as histogram and as summary).
* Number of requests to the service.
