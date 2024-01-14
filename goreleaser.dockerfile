# Final harded image from scratch
# Build from google distroless projekt image
# FROM gcr.io/distroless/static AS final 
FROM ubuntu:22.04 AS final
COPY golang-custom-rpi-exporter /golang-custom-rpi-exporter

USER nobody
EXPOSE 8080
VOLUME ["/tmp", "/var/run", "/sys/class/thermal/thermal_zone0/temp", "/var/lib/apt/lists", "/var/lib/dpkg"]
CMD [ "/golang-custom-rpi-exporter" ]