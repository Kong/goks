#!/bin/bash -e

GOLANGCI_LINT_VERSION=v1.45.0
curl -sSfL \
  "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh" \
  | sh -s -- -b bin ${GOLANGCI_LINT_VERSION}