#!/bin/bash

# Build the Go binary and package it into a par file
# This is designed to be run from the root of the project
go build ./cmd/blank
wash par create --vendor wasmcloud --name "Blank Go" --binary ./blank --compress
rm blank