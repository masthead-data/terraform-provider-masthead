# Contributing

Thanks for your interest in contributing to the Masthead Data Terraform Provider.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.25

## Developing the Provider

1. Build the provider:

   ```shell
   go mod tidy
   go install
   ```

   This will put the provider binary in the `$GOPATH/bin` directory.

2. Update the documentation:

   ```shell
   make generate
   ```

3. Add the following block to your Terraform configuration in `~/.terraformrc` to use the provider from your local development environment:

   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/masthead-data/masthead" = "<PATH>"
     }
     # For all other providers, install them directly from their origin provider
     # registries as normal. If you omit this, Terraform will _only_ use
     # the dev_overrides block, and so no other providers will be available.
     direct {}
   }
   ```

4. Run the test resources deployment:

   ```shell
   terraform -chdir=examples/provider init
   terraform -chdir=examples/provider plan -var api_token=YOUR_API_TOKEN
   ```

## Developing the Go client

For local development of the Masthead Data client library, you can use the following steps to set up your environment:

1. Build the client by running:

   ```shell
   go mod tidy
   go install
   ```

2. Run `make testacc` to run the acceptance tests.

   *Note:* Acceptance tests create real resources.

## Ways to Contribute

- Report bugs and edge cases.
- Propose or implement new features.
- Improve tests, examples, and documentation.

## Development Guidelines

- Follow [Terraform Plugin Framework guidance](https://developer.hashicorp.com/terraform/plugin/framework).
- Keep commit messages clear and concise.
- Add tests for new features and bug fixes.
- Update documentation when behavior or schema changes.
- Ensure tests pass before opening a pull request.
- Keep feedback respectful and constructive.

## Local Development

1. Build and install:

   ```shell
   go mod tidy
   go install
   ```

2. Regenerate docs:

   ```shell
   make generate
   ```

3. Run tests:

   ```shell
   make test
   ```

4. Run acceptance tests (requires credentials):

   ```shell
   make testacc
   ```
