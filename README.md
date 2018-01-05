marionette
==========

This is a Go port of the [marionette][] programmable networy proxy.

## Development

When adding new formats, you'll need to first install `go-bindata`:

```sh
$ go get -u github.com/jteeuwen/go-bindata/...
```

Then you can use `go generate` to convert the asset files to Go files:

```sh
$ go generate ./...
```



