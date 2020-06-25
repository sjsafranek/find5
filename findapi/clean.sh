#!/bin/bash

files=$(find lib/ -type f -name "*.go")
for file in $files; do
    go fmt $file
done
