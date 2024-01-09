# This docker files does a multi stage docker build and creates
# a statically cross platform compiled binary shipped in a harded
# docker image from scratch

ARG GO_VERSION="1.21"
ARG ALPINE_VERSION="3.19"

# Golang cross compiling statical binary
# FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build

ARG TARGETOS
ARG TARGETARCH
ARG APP_VERSION

WORKDIR /build

COPY ["go.mod", "go.sum", "./"]
RUN go mod download

# COPY ["*.go", "./"]
COPY . . 

# ldflags -w disables debug, letting the file be smaller.
# netgo makes sure we use built-in net package and not the system’s one.
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS}  \
    GOARCH=${TARGETARCH}  \
    go build -a -tags netgo -ldflags "-X main.version=${APP_VERSION} -w -extldflags '-static'" -o /golang-custom-rpi-exporter

# Final harded image from scratch
# Build from google distroless projekt image
# FROM gcr.io/distroless/static AS final 
FROM ubuntu:22.04 AS final
COPY --from=build /golang-custom-rpi-exporter /golang-custom-rpi-exporter

USER nobody
EXPOSE 8080
VOLUME ["/tmp", "/var/run", "/sys/class/thermal/thermal_zone0/temp", "/var/lib/apt/lists", "/var/lib/dpkg"]
CMD [ "/golang-custom-rpi-exporter" ]