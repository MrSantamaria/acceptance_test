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
RUN go build -o /app/acceptance-test

### STAGE 3: Final ###
FROM quay.io/openshift/origin-tools:latest

# Copy the Go binary from the build stage to the final image
COPY --from=build /app/acceptance-test .

# Creates the directory for the ocm config file
RUN mkdir -p /.config/ocm && chmod -R 775 /.config

# Run the Go application when the container starts
CMD ["./acceptance-test"]