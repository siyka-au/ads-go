# Testing

First be sure to set your ADS target NetId
```pwsh
$env:ADS_TARGET_NET_ID = "199.4.42.250.1.1"
```

- Run all unit tests across the repo (excludes integration build-tag tests):
```pwsh
go test ./...
```

- Run all tests (including tests with the integration build tag):
```pwsh
go test -v -tags=integration ./...
```

- Run all tests in the integration package:
```pwsh
go test -v -tags=integration ./test/integration
```

- Run a single top-level test in the integration package:
```pwsh
go test -v -tags=integration ./test/integration -run '^TestReadAllAccessPaths$'
```

- Run a specific subtest (e.g., sint_max inside TestStaticSeed):
```pwsh
go test -v -tags=integration ./test/integration -run '^TestStaticSeed/sint_max$'
```

- List all tests in the package:
```pwsh
go test -list . -tags=integration ./test/integration
```
