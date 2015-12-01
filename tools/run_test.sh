#!/bin/sh

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

docker run --rm -it --env-file $DIR/../.env -v $DIR/../:/data -w /data --link spul_connector golang:1.5.1 go run $1
