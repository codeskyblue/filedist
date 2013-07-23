## lwget
base on wget, offer md5sum output when finished downloading.

## usage
    lwget [OPTIONS]  <URL> <target>

    Help Options:
      -h, --help=       Show this help message

    Application Options:
      -t, --timeout=    down timeout (0s)
      -l, --limit-rate= download speed limit per second (10m)
      -m, --md5sum=     check if md5sum matches
          --wget=       specfity which wget to use (/usr/bin/wget)

## example
### normal download, and limit rate to 10m/s
    lwget --limit-rate=10m http://code.jquery.com/jquery-1.10.1.min.js jquery.min.js

will save as jquery.min.js
### add md5sum check

    lwget --md5sum=12334234564574562 http://code.jquery.com/jquery-1.10.1.min.js jquery.min.js

this is an error md5sum, and this will output, exit !0

    expect (12334234564574562) but got (33d85132f0154466fc017dd05111873d)

### add timeout

    lwget -l 1m --timeout=2s http://www.ubuntu.com/start-download?distro=desktop&bits=64&release=lts ubuntu.iso

this will raise timeout. exit !0

