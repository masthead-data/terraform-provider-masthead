# Masthead Data Terraform Provider

This repository is a [Terraform](https://www.terraform.io) provider for Masthead Data.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.23

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To update the dependencies and build the provider, run:
    ```shell
    go mod tidy
    go install
    ```

This will put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

## Developing the GO client

For local developement of the Masthead Data client library, you can use the following steps to set up your environment:

1. Adjust the `go.mod` file to point to the local version of the Masthead Data client library. This is done by replacing the `masthead` module with a local path. Use the `replace` directive in your `go.mod` file to point to the local path of the Masthead Data client library.

    ```go
    require masthead v0.0.0
    replace masthead => ../../internal/client
    ```

2. Import the local version of the Masthead Data client library in your code.

    ```go
    import (
        ...
        "masthead"
    )
    ```

3. Build the client by running:

    ```shell
    go mod tidy
    go install
    ```

to ensure that all dependencies are correctly resolved and build the API client with the updated client libraries.
4. Run `make testacc` to run the acceptance tests with the local client library.

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
