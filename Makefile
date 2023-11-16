COVER_PROFILE = /tmp/go-cover.tmp
COVER_OPTION = -coverprofile "${COVER_PROFILE}"

# code check
code-check:
	golangci-lint run

# tests
test:
	go test ./...

# test coverage
coverage:
	go test ./... ${COVER_OPTION}

# test coverage in html
coverage-html: coverage
	go tool cover -html="${COVER_PROFILE}"
