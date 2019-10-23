#!/bin/bash

go mod download
go build -o main
./main
