# STANDALONE EXECUTION

You can start this example just by running:

`go run main.go`

If you go to `localhost:8080`, you'll see log traces in your console.
You can also go to `localhost:8282/traces` to have a look at your traces in `Appdash`.

# EXECUTION WITH PROMETHEUS

If you want to test how Prometheus metrics work, run:

`make deps && make up && go run main.go`

Then open `localhost:9090/graph` and execute `lzd_request_counter`. Don't forget to send some requests to `localhost:8080`; each request should generate `3` subrequests.
