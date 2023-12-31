### STAGE 1: Build ###
FROM registry.ci.openshift.org/openshift/release:golang-1.19 AS build

ENV GOFLAGS=
ENV PKG=/go/src/github.com/openshift/acceptance-test
WORKDIR ${PKG}

# Install needed binaries
RUN go install github.com/openshift-online/ocm-cli/cmd/ocm@latest
RUN go install github.com/observatorium/obsctl@main

# Copy the entire directory into the container
COPY . .

# Print the Go environment variables
RUN go env

# Build the Go application
RUN go build -o /acceptance-test

### STAGE 2: Final ###
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Copy the Go binary from the build stage to the final image
COPY --from=build /acceptance-test /acceptance-test

# Copy the binaries from Stage 1 to /usr/local/bin in the final image
COPY --from=build /go/bin/ocm /usr/local/bin/ocm
COPY --from=build /go/bin/obsctl /usr/local/bin/obsctl

# Set the working directory in the final image
WORKDIR /

# Give execution permissions
RUN chmod +x /acceptance-test
RUN chmod +x /usr/local/bin/ocm
RUN chmod +x /usr/local/bin/obsctl

# Run the Go application when the container starts
CMD ["./acceptance-test"]
