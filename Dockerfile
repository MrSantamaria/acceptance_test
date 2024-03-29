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

### STAGE 1: Build ###
# ... [your existing Stage 1 content] ...

### STAGE 2: Final ###
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Install shadow-utils to get the useradd command
RUN microdnf install shadow-utils

# Create a directory for your application and set proper permissions
RUN mkdir /app && chmod 755 /app

# Create the .config directory at the root and set permissions
RUN mkdir /.config && chmod 777 /.config

# Copy the Go binary from the build stage to the final image
COPY --from=build /acceptance-test /app/acceptance-test

# Give execution permissions
RUN chmod +x /app/acceptance-test

# Copy the binaries from Stage 1 to /usr/local/bin in the final image
COPY --from=build /go/bin/ocm /usr/local/bin/ocm
COPY --from=build /go/bin/obsctl /usr/local/bin/obsctl

# Give execution permissions to copied binaries
RUN chmod +x /usr/local/bin/ocm
RUN chmod +x /usr/local/bin/obsctl

# Create a non-root user and group
RUN useradd -u 1001030000 -r -g 0 -d /app -s /sbin/nologin -c "Default Application User" default

# Change the ownership of the /app directory and /.config to the non-root user
RUN chown -R 1001030000:0 /app /.config

# Switch to the non-root user
USER 1001030000

# Set the working directory in the final image
WORKDIR /app

# Run the Go application when the container starts
CMD ["./acceptance-test"]
