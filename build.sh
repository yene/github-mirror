#!/bin/bash
set -e
go build -o github-mirror
upx github-mirror
