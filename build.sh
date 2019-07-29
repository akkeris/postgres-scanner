#!/bin/sh

cd /go/src
go get        "github.com/lib/pq"

cd /go/src/postgres-scanner
go build postgres-scanner.go

