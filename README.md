marionette
==========

This is a Go port of the [marionette][] programmable networy proxy.

## Installation

Marionette requires several dependencies to be installed first.

### GMP

Download the latest version of [GMP][], unpack the
archive and run:

```sh
$ ./configure --enable-cxx
$ make
$ sudo make install
$ make check
```



### PyCrypto

Download the latest version of [PyCrypto][], unpack the archive and run:

```sh
# https://ftp.dlitz.net/pub/dlitz/crypto/pycrypto/pycrypto-2.6.1.tar.gz

$ python setup.py build
$ sudo python setup.py install --user
```


### regex2dfa

Download the latest version of [regex2dfa][], unpack the archive and run:

```sh
# https://github.com/kpdyer/regex2dfa/archive/master.zip

$ ./configure
$ make
$ sudo python setup.py install --user
```


### libfte

Download the latest version of [libfte][], unpack the archive and run:

```sh
# https://github.com/kpdyer/libfte/archive/master.zip

$ sudo python setup.py install --user
```


[marionette]: https://github.com/marionette-tg/marionette
[GMP]: https://gmplib.org
[PyCrypto][]: https://www.dlitz.net/software/pycrypto/
[regex2dfa]: https://github.com/kpdyer/regex2dfa/archive/master.zip



## Development

When adding new formats, you'll need to first install `go-bindata`:

```sh
$ go get -u github.com/jteeuwen/go-bindata/...
```

Then you can use `go generate` to convert the asset files to Go files:

```sh
$ go generate ./...
```

To install the original [marionette][] library for comparing tests, download
the latest version, unpack the archive and run:

### Marionette

```sh
$ python setup.py install
```


## Testing

Use the built-in go testing command to run the unit tests:

```sh
$ go test ./...
```

To run integration tests, specify the `integration` tag:

```sh
$ go test -tags integration ./...
```