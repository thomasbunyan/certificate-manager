# Certbot Lambda

An AWS Lambda used to issue, track and refresh TLS certificates provided by [Let’s Encrypt](https://letsencrypt.org/).

## Infrastructure

![architecture_diagram](./assets/arch.svg)

## Usage

### Prerequisites

- A registered domain name
- DNS configuration hosted in Route53

### Adding/removing a domain

_todo_

## Development

#### Go clients

https://letsencrypt.org/docs/client-options/#libraries-go

https://go-acme.github.io/lego/

#### JSON Web Key

https://www.ietf.org/rfc/rfc7517.txt

## Author

[Thomas Bunyan](https://github.com/thomasbunyan)

github.com/vittorio-nardone/certbot-lambda

## License

Copyright © 2022 [Thomas Bunyan](https://github.com/thomasbunyan).

Usage is provided under the MIT License.
