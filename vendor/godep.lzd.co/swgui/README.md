# Swagger UI

Package swgui (Swagger UI) provide a HTTP handler to serve Swagger UI.
All assets are embedded in GO source code, so just build and run.

## How to use

```go
    package main

    import "http"
    import "godep.lzd.co/swgui"

    func main() {
        http.Handle("/", swgui.NewHandler("Page title", "path/to/swagger.json", "/"))
        http.ListenAndServe(":8080", nil)
    }
```

## Run as standalone server

Install swgui server

    go get godep.lzd.co/swgui/...

Start server

    swgui-server -port 8080

## GoDoc 

[![GoDoc](https://godoc.org/bitbucket.org/lazadaweb/swgui?status.svg)](https://godoc.org/bitbucket.org/lazadaweb/swgui)

    godoc godep.lzd.co/swgui
    
## Updating

 * Clone and update LZD fork of Swagger UI `https://gitlab.lzd.co/vpoturaev/swagger-ui-lzd`
 * Build with 
```
npm install
gulp
```
 * Merge files from `./dist` there to `./static` here
 * `go generate`
