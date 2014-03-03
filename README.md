# Install `github.com/hcatlin/libsass`

1.

create folder : `$GOPATH/clibs/include`

2. Don't use "go get" to install this package, because gosass.go : `#cgo LDFLAGS: -L../../clibs/lib -lsass -lstdc++`

    cd $GOAPTH/src
    git clone https://github.com/moovweb/gosass.git

3.

    go get github.com/hcatlin/libsass
    cp $GOPATH/src/github.com/hcatlin/libsass/*.h $GOPATH/clibs/include/



