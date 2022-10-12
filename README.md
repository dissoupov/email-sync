# email-sync

GoLang library for syncing email accounts.

## Usage 

TODO:

## CLI

1. To use CLI locally, create `~/.ableai/.email.yaml` file, and populate `client_key`, `client_secret` fields with real values.

```yaml
---
login_server: https://test.salesforce.com
proxy_server:
http_timeout: 30s
api_ver: v51.0
user: xxx@ableai.com.ncinodev1
password: xxx
token_cache: ~/.ableai/.token.json
oauth:
  client_key: 3MVxxx
  client_secret: D2xxx
  expiry: 30m
```

```sh
Usage: email-sync <command>

TODO:
```

## Requirements

1. GoLang 1.18+

You can check your GoLang version by running `go version`

## Contribution

* `make all` complete build and test
* `make test` run the tests
* `make testshort` runs the tests skipping the end-to-end tests and the code coverage reporting
* `make covtest` runs the tests with end-to-end and the code coverage reporting
* `make coverage` view the code coverage results from the last make test run.
* `make generate` runs go generate to update any code generated files
* `make fmt` runs go fmt on the project.
* `make lint` runs the go linter on the project.

run `make all` once, then run `make build` or `make test` as needed.

First run:

    make all

Tests:

    make test

Optionally run golang race detector with test targets by setting RACE flag:

    make test RACE=true

Review coverage report:

    make covtest coverage

