FROM alpine:3.23.0 AS builder
ARG TARGETOS
ARG TARGETARCH

# Install dependencies.
RUN apk add --no-cache \
    nodejs \
    npm 

WORKDIR /app
# Copy packge.json and lock file.
COPY package.json package.json
COPY package-lock.json package-lock.json
# Cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer.
# Use clean npm install: npm ci.
RUN npm ci

# copy source code (relies on .dockerignore to filter)
# or you could clone the git repo
COPY . .

## build
# The build has no default value to allow npm to be built according to the host where the command was called. 
# For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container will have the same platform.
RUN echo "Building for OS: $TARGETOS, ARCH: $TARGETARCH" && npm run build

# Use a minimal base image to package the code.
# For a binary i would preffere a distroles image like: gcr.io/distroless/static:nonroot.
# Refer to https://github.com/GoogleContainerTools/distroless for more details.
FROM nginxinc/nginx-unprivileged:alpine3.22-perl

USER root
# Update packages if you need to i would recommend specifing the software and version.
# This is just for testing.
RUN apk update && apk upgrade --no-cache

## Copy your code after the npm build.
COPY --from=builder /app/dist /usr/share/nginx/html

# Expose webserver port: This is the default port for our nginx rootless container.
# Check wich port ur webserver does listen on.
EXPOSE 8080

# We use the nonroot user.
# nginx user = 101:101
USER nginx

# Now start nginx.
CMD ["nginx", "-g", "daemon off;"]
