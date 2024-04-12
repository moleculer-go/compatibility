# compatibility-tests

Compatibility tests with other moleculer implementations.

[![Build Status](https://cloud.drone.io/api/badges/moleculer-go/compatibility/status.svg)](https://cloud.drone.io/moleculer-go/compatibility)

## Moleculer JS

## dependencies

YOu need Node JS with Npm installed

The test will install the moleculerJS and its dependencies and start moleculer JS using node for testing.

For the NATS test you need NATS running:

```
docker run -d -p 4222:4222 nats-streaming -mc 0
```

Test runners:

```
go get github.com/onsi/ginkgo/ginkgo/outline@v1.16.2
go get github.com/onsi/ginkgo/ginkgo@v1.16.2
```

## Running tests

```
go run github.com/onsi/ginkgo/ginkgo -r
```
