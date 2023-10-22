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

### STAGE 2: Grab oc binary ###
FROM alpine:latest AS downloader

WORKDIR /tmp

# Use wget to download the tar.gz file
RUN wget -O openshift-client-linux.tar.gz https://mirror.openshift.com/pub/openshift-v4/clients/ocp-dev-preview/latest/openshift-client-linux.tar.gz
RUN wget -O ocm https://github.com/openshift-online/ocm-cli/releases/download/v0.1.70/ocm-linux-amd64

# Untar the downloaded tar.gz file
RUN tar -xzvf openshift-client-linux.tar.gz

### STAGE 3: Final ###
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Copy the Go binary from the build stage to the final image
COPY --from=build /acceptance-test .

# Copy the extracted binary from Stage 2 to /usr/local/bin in the final image
COPY --from=downloader /tmp/oc /usr/local/bin/oc
COPY --from=downloader /tmp/ocm /usr/local/bin/ocm

RUN chmod +x /acceptance-test
RUN chmod +x /usr/local/bin/oc
RUN chmod +x /usr/local/bin/ocm
RUN mkdir /.vscode-server
RUN mkdir /.vscode-server-insiders
RUN chmod 777 /.vscode-server
RUN chmod 777 /.vscode-server-insiders

# Run the Go application when the container starts
CMD ["./acceptance-test"]