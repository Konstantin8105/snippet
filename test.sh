#!/bin/bash

cp ./testdata/cli.actual ./testdata/cli.actual.1
go run gosnippet.go -w -e="./testdata/cli.expect" ./testdata/cli.actual.1
diff -sq ./testdata/cli.actual.1 ./testdata/cli.result
rm -f ./testdata/cli.actual.1