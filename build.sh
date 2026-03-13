#!/usr/bin/bash

set -euo pipefail

VERSION="${VERSION:-dev}"

GOOS=windows GOARCH=amd64 go build -ldflags "-X hjbdev/pvm/commands.version=${VERSION}" -o pvm.exe .
