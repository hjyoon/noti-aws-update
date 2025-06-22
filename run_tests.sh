#!/bin/sh

set -eu

echo "==== [ go test -v -cover ./... ] ===="
go test -v -cover ./...

echo
echo "==== [ go test -coverprofile=coverage.out ./... ] ===="
go test -coverprofile=coverage.out ./...

echo
echo "==== [ go tool cover -func=coverage.out ] ===="
go tool cover -func=coverage.out

# 커버리지 HTML 리포트 자동 오픈 (옵션)
if command -v xdg-open >/dev/null 2>&1; then
    go tool cover -html=coverage.out -o coverage.html
    echo "==== [ Open coverage.html (Linux) ] ===="
    xdg-open coverage.html
elif command -v open >/dev/null 2>&1; then
    go tool cover -html=coverage.out -o coverage.html
    echo "==== [ Open coverage.html (macOS) ] ===="
    open coverage.html
else
    echo "==== [ For full HTML report: run ] ===="
    echo "go tool cover -html=coverage.out"
fi
