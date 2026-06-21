module github.com/davidnguyen2205/opero/backend

go 1.25.7

// Dependencies are resolved by `go mod tidy` (run `make tidy`).
// Expected direct deps once tidied:
//   github.com/go-chi/chi/v5
//   github.com/jackc/pgx/v5

require (
	github.com/getkin/kin-openapi v0.140.0
	github.com/go-chi/chi/v5 v5.3.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.10.0
	github.com/oapi-codegen/runtime v1.4.2
	github.com/pressly/goose/v3 v3.27.1
	golang.org/x/crypto v0.53.0
)

require (
	github.com/apapsch/go-jsonmerge/v2 v2.0.0 // indirect
	github.com/go-openapi/jsonpointer v0.22.5 // indirect
	github.com/go-openapi/swag/jsonname v0.25.5 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/oasdiff/yaml v0.1.0 // indirect
	github.com/oasdiff/yaml3 v0.0.13 // indirect
	github.com/rogpeppe/go-internal v1.15.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	github.com/sethvargo/go-retry v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sync v0.21.0 // indirect
	golang.org/x/text v0.38.0 // indirect
)
