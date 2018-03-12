marionette
==========

This is a Go port of the [marionette][] programmable networy proxy.

## WebBrowser Demonstration

Please install Marionette as described below, and then go to the web browser demonstration page [here](./BrowserDemo.md)

## Installation

Marionette requires several dependencies to be installed first.

### GMP

Download the latest version of [GMP][], unpack the
archive and run:

```sh
$ wget https://gmplib.org/download/gmp/gmp-6.1.2.tar.bz2
$ tar -xvjf gmp-6.1.2.tar.bz2
$ cd gmp-6.1.2

$ ./configure --enable-cxx
$ make
$ sudo make install
$ make check
```



### PyCrypto

Download the latest version of [PyCrypto][], unpack the archive and run:

```sh
$ wget https://ftp.dlitz.net/pub/dlitz/crypto/pycrypto/pycrypto-2.6.1.tar.gz
$ tar zxvf pycrypto-2.6.1.tar.gz
$ cd pycrypto-2.6.1

$ python setup.py build
$ sudo python setup.py install --user
```


### regex2dfa

Download the latest version of [regex2dfa][], unpack the archive and run:

```sh
$ wget -O regex2dfa.zip https://github.com/kpdyer/regex2dfa/archive/master.zip
$ unzip regex2dfa.zip
$ cd regex2dfa-master

$ ./configure
$ make
$ sudo python setup.py install --user
```


### libfte

Download the latest version of [libfte][], unpack the archive and run:

```sh
$ wget -O libfte.zip https://github.com/kpdyer/libfte/archive/master.zip
$ unzip libfte.zip
$ cd libfte-master

$ sudo python setup.py install --user
```


### Building the Marionette Binary

First, make sure you have installed Go from [https://golang.org/][go]. Next,
install `dep` using [these instructions][dep].

Finally, retrieve the source, update project dependencies, and install the
`marionette` binary:

```sh
$ go get github.com/redjack/marionette
$ cd $GOPATH/src/github.com/redjack/marionette
$ dep ensure
$ go install ./cmd/marionette
```

The `marionette` binary is now installed in your `$GOPATH/bin` folder.


[marionette]: https://github.com/marionette-tg/marionette
[GMP]: https://gmplib.org
[PyCrypto]: https://www.dlitz.net/software/pycrypto/
[regex2dfa]: https://github.com/kpdyer/regex2dfa/archive/master.zip
[libfte]: https://github.com/kpdyer/libfte
[go]: https://golang.org/
[dep]: https://github.com/golang/dep#installation



## Installation (Docker)

You can also install `marionette` as a Docker image. You'll need to have Docker
installed. You can find instructions for specific operating system here:
https://docs.docker.com/install

The easiest way to setup `marionette` is to use the provided `Dockerfile`.
First, build your docker image. Please note that this can take a while.

```
$ docker build -t redjack/marionette:latest .
```


### Running using the Docker image

Next, run the Docker image and use the appropriate port mappings for the
Marionette format you're using. For example, `http_simple_blocking` uses
port `8081`:

```sh
$ docker run -p 8081:8081 redjack/marionette server -format http_simple_blocking
```

```sh
$ docker run -p 8079:8079 redjack/marionette client -bind 0.0.0.0:8079 -format http_simple_blocking
```

If you're running _Docker for Mac_ then you'll also need to add a `-server` argument:

```sh
$ docker run -p 8079:8079 redjack/marionette client -bind 0.0.0.0:8079 -server docker.for.mac.host.internal -format http_simple_blocking
```


### Using a fixed `channel.bind()` port

The `ftp_simple_blocking` uses randomized ports for the `channel.bind()` plugin.
Unfortunately, `docker` does not support random port mappings so you can
hardcode the port using the `MARIONETTE_CHANNEL_BIND_PORT` environment variable:

```sh
$ docker run -p 2121:2121 -e MARIONETTE_CHANNEL_BIND_PORT='6000' redjack/marionette server -format ftp_simple_blocking
```

```sh
$ docker run -p 8079:8079 redjack/marionette client -format ftp_simple_blocking
```



## Demo

### HTTP-over-FTP

In this example, we'll mask our HTTP traffic as FTP packets.

First, follow the installation instructions above on your client & server machines.

Start the server proxy on your server machine and forward traffic to a server
such as `google.com`.

```sh
$ marionette server -format ftp_simple_blocking -proxy google.com:80
listening on [::]:2121, proxying to google.com:80
```

Start the client proxy on your client machine and connect to your server proxy.
Replace `$SERVER_IP` with the IP address of your server.

```sh
$ marionette client -format ftp_simple_blocking -server $SERVER_IP
listening on 127.0.0.1:8079, connected to <SERVER_IP>
```

Finally, send a `curl` to `127.0.0.1:8079` and you should see a response from
`google.com`:

```sh
$ curl 127.0.0.1:8079
```


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