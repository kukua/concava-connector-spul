# Installation

The SPUL connector can be run as a Golang program or in a Docker container.

Make sure [ConCaVa](https://github.com/kukua/concava) is setup aswell.
See [`.env.sample`](https://github.com/kukua/concava-spul-connector/tree/master/.env.sample) for the default configuration.

## Golang

```bash
git clone https://github.com/kukua/concava-spul-connector.git
cd concava-spul-connector
cp .env.sample .env
# > Edit .env

source .env
go run src/connector.go
```

Tested with Go v1.5.1.

## Docker

First, [install Docker](http://docs.docker.com/engine/installation/). Then run:

```bash
curl https://raw.githubusercontent.com/kukua/concava-spul-connector/master/.env.sample > .env
# > Edit .env

docker run -d -p 3333 -p 5555 --env-file .env --name spul_connector kukuadev/concava-spul-connector
```

Tested with Docker v1.8.
