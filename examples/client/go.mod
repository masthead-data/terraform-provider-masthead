module github.com/masthead-data/terraform-provider-masthead/examples/client

go 1.24.1

require (
	github.com/masthead-data/terraform-provider-masthead/internal/client v0.1.0
)

// for local testing
//require masthead v0.0.0
//replace masthead => ../../internal/client
