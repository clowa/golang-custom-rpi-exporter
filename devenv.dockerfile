FROM ubuntu:22.04

RUN apt update && apt install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# RUN touch /var/run/reboot-required