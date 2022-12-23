# Certificate manager Lambda

An AWS Lambda used to issue, track and refresh TLS certificates provided by [Let’s Encrypt](https://letsencrypt.org/).

## Infrastructure

![architecture_diagram](./assets/arch.svg)

## Usage

### Prerequisites

- A registered domain name
- DNS configuration hosted in Route53

## Development

### Go clients

https://letsencrypt.org/docs/client-options/#libraries-go

https://go-acme.github.io/lego/

## Author

[Thomas Bunyan](https://github.com/thomasbunyan)

## License

Copyright © 2022 [Thomas Bunyan](https://github.com/thomasbunyan).

Usage is provided under the MIT License.
