# Testing

Make sure a SPUL container is running with the name 'spul_connector'. Then run:

```js
./tools/run_test.sh tools/test_timestamp.go
./tools/run_test.sh tools/test_spul.go
docker logs spul_connector
```
