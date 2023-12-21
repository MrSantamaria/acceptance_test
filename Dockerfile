### STAGE 1: Build ###
FROM registry.ci.openshift.org/openshift/release:golang-1.19 AS build

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/acceptance-test
WORKDIR ${PKG}

# Copy the entire directory into the container
COPY . .

# Print the Go environment variables
RUN go env

# Build the Go application
RUN go build -o /acceptance-test

### STAGE 2: Download needed binaries ###
FROM alpine:latest AS downloader

WORKDIR /tmp

# Use wget to download the tar.gz file
RUN go install github.com/openshift-online/ocm-cli/cmd/ocm@latest
RUN go install github.com/observatorium/obsctl@main

# Untar the downloaded tar.gz file
RUN tar -xzvf openshift-client-linux.tar.gz

### STAGE 3: Final ###
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Copy the Go binary from the build stage to the final image
COPY --from=build /acceptance-test .

# Copy the extracted binary from Stage 2 to /usr/local/bin in the final image
COPY --from=downloader /go/bin/ocm /usr/local/bin/ocm
COPY --from=downloader /go/bin/obsctl /usr/local/bin/obsctl

RUN chmod +x /acceptance-test
RUN chmod +x /usr/local/bin/ocm
RUN chmod +x /usr/local/bin/obsctl

# Run the Go application when the container starts
CMD ["./acceptance-test"]