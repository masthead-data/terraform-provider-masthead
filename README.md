# Masthead Data Terraform Provider

This repository is a [Terraform](https://www.terraform.io) provider for Masthead Data.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

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
        "hashicorp.com/edu/hashicups" = "<PATH>"
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

## Developing the GO client

For local developement of the Masthead Data client library, you can use the following steps to set up your environment:

1. Build the client by running:

    ```shell
    go mod tidy
    go install
    ```

2. Run `make testacc` to run the acceptance tests.

    *Note:* Acceptance tests create real resources.

## Contributing

We welcome contributions to the Masthead Data Terraform Provider! If you have a bug fix, feature request, or improvement, please open an issue or pull request.
Please ensure that your code adheres to the following guidelines:

- Follow the [Terraform Provider Development Guidelines](https://www.terraform.io/docs/plugin-sdkv2/).
- Write clear and concise commit messages.
- Add tests for new features or bug fixes.
- Update documentation as needed.
- Ensure that all tests pass before submitting your pull request.
- Please be respectful and constructive in your feedback.

Thank you for contributing to the project!
