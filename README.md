# Docker machine driver for [Hetzner Cloud](https://www.hetzner.com/cloud)

[![Go Report Card](https://goreportcard.com/badge/github.com/eduardnikolenko/docker-machine-driver-hetzner)](https://goreportcard.com/report/github.com/eduardnikolenko/docker-machine-driver-hetzner)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](/LICENSE)

## Installation

Use go get github.com/eduardnikolenko/docker-machine-driver-hetzner and make sure that docker-machine-driver-hetzner is located somwhere in your PATH

## Usage

    $ docker-machine create \
      --driver hetzner \
      --hetzner-access-token=<YOU_ACCESS_TOKEN> \
      my-machine

## Options

| Parameter                    | Env                    | Default                      |
| ---------------------------- | ---------------------- | ---------------------------- |
| **`--hetzner-access-token`** | `HETZNER_ACCESS_TOKEN` | -                            |
| **`--hetzner-image`**        | `HETZNER_IMAGE`        | `debian-9`                   |
| **`--hetzner-location`**     | `HETZNER_LOCATION`     | `fsn1`                       |
| **`--hetzner-server-type`**  | `HETZNER_SERVER_TYPE`  | `cx11`                       |

## License

MIT © [Eduard Nikolenko](https://github.com/eduardnikolenko)
