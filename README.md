# SPUL Connector

> SPUL Connector for converting sensor data to a binary payload and forwarding it to ConCaVa.

## How to use

```bash
cp .env.sample .env
# > Edit .env
docker run -d -p 3333 -p 5555 --env-file .env --name spul_connector kukuadev/concava-spul-connector
```

Make sure [ConCaVa](https://github.com/kukua/concava) is setup aswell.
See [`.env.sample`](https://github.com/kukua/concava-spul-connector/tree/master/.env.sample) for the default configuration.

## Test

```js
./tools/run_test.sh tools/test_timestamp.go
./tools/run_test.sh tools/test_spul.go
```

## Contribute

Your help and feedback are highly welcome!
