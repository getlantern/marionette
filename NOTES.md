NOTES
=====

## Python Marionette Installation

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

## Development

To install the original [marionette][] library for comparing tests, download
the latest version, unpack the archive and run:

### Marionette

```sh
$ python setup.py install
```


[marionette]: https://github.com/marionette-tg/marionette
[GMP]: https://gmplib.org
[PyCrypto][]: https://www.dlitz.net/software/pycrypto/
[regex2dfa]: https://github.com/kpdyer/regex2dfa/archive/master.zip
