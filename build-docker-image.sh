#!/usr/bin/env sh
set -eu

docker build --build-arg GOPROXY=https://goproxy.cn,direct -t mtit/anox-gateway:latest .