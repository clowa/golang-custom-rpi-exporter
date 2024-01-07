[![Build docker images](https://github.com/clowa/golang-custom-rpi-exporter/actions/workflows/docker-buildx.yaml/badge.svg)](https://github.com/clowa/golang-custom-rpi-exporter/actions/workflows/docker-buildx.yaml)

# Overview

This node exporter is supposed as an addition to the [official node exporter](https://github.com/prometheus/node_exporter) for Raspberry Pi. It exposes some metrics about the Raspberry Pi itself and `apt` packages.

You can find the docker images on [Docker Hub](https://hub.docker.com/r/clowa/golang-custom-rpi-exporter).

Supported platforms:

- `linux/arm`
- `linux/arm64`

# Getting started

# To-Do

- [ ] Improve performance. Currently the exporter takes about ~4 seconds to scrape all metrics. Perhaps due to Filesystem i/o.
- [ ] Add more metrics
  - [ ] Subset of information from [`vcgencmd`](https://www.raspberrypi.com/documentation/computers/os.html#vcgencmd)
