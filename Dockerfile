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

### STAGE 2: Grab oc binary ###
FROM image-registry.openshift-image-registry.svc:5000/openshift/cli:latest as oc

### STAGE 3: Final ###
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Copy the Go binary from the build stage to the final image
COPY --from=build /app/acceptance-test .

# Copy the oc binary from the build stage to the final image
COPY --from=oc /usr/bin/oc /usr/bin/oc

# Creates the directory for the ocm config file
RUN mkdir -p /.config/ocm && chmod -R 775 /.config

# Run the Go application when the container starts
CMD ["./acceptance-test"]