#Pre-build stage
FROM golang:1.21-bookworm AS pre-build
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir "/build"

WORKDIR "/build/"
# Get dependencies in a layer so rebuilds are faster
COPY go.mod . 
COPY go.sum .
RUN go mod download

# Build the base application
COPY . /build/
RUN go build ./cmd/mycolog

#just in case, make the binaries executable
RUN <<EOT bash 
    chmod +x /build/mycolog
EOT

# post copying / run stage, should only rerun if scripts or bins change
FROM debian:12-slim
SHELL ["/bin/bash", "-o", "pipefail", "-c"]
RUN mkdir "/scripts" && mkdir "/temp"
COPY --from=pre-build "/build/mycolog" "/usr/bin/"
COPY "mycolog.sh" "/scripts"
COPY "docker-entrypoint.sh" "/scripts"

WORKDIR "/scripts"
ENTRYPOINT ["/bin/bash", "./docker-entrypoint.sh"]
CMD ["./mycolog.sh"]
