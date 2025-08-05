# Accept the Go version for the image to be set as a build argument.
# Default to Go 1.24
ARG GO_VERSION=1.24

# First stage: build the executable.
FROM public.ecr.aws/docker/library/golang:${GO_VERSION}-alpine AS builder

# Create the user and group files that will be used in the running container to
# run the process as an unprivileged user.
RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

# Set the working directory outside $GOPATH to enable the support for modules.
WORKDIR /src

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build
COPY ./go.mod ./go.sum ./
RUN go mod tidy
RUN go mod download

# Import the code from the context.
COPY ./ ./

# Build the executable to `/app`. Mark the build as statically linked.
RUN CGO_ENABLED=0 go build \
    -installsuffix 'static' \
    -o /app ./cmd/main.go

COPY ./.env /.env

# Final stage: the running container.
FROM public.ecr.aws/docker/library/alpine:latest AS final

# Install debugging tools
RUN apk add --no-cache bash curl vim net-tools

# Import the user and group files from the first stage.
COPY --from=builder /user/group /user/passwd /etc/
# copy .env
COPY --from=builder /.env /.env

# Import the compiled executable from the first stage.
# COPY --from=builder . .
COPY --from=builder /app /app

# Declare the port on which the inventory service will be exposed.
ENV PORT=4000

EXPOSE ${PORT}

# Perform any further action as an unprivileged user.
USER nobody:nobody

# Run the compiled binary.
ENTRYPOINT ["/app"]