# Testing

First, add `spul_connector` to your hosts file. Pointing to either localhost (`127.0.0.1`) or the Docker container IP. Then run:

```js
./tools/run_test.sh tools/test_timestamp.go
./tools/run_test.sh tools/test_spul.go
```
