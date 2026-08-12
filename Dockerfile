# FARO transparency log image.
#
# Multi-stage, ending at distroless/static as a non-root user: the runtime layer
# has no shell, no package manager and no libc.
#
# The image ships only the log service. The audit explorer in web/ is a static
# Next.js app and is deployed separately, behind a CDN; keeping them apart means
# a compromise of the public website cannot reach the log's signing key.

FROM golang:1.26-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/faro-log ./cmd/faro-log

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/faro-log /usr/local/bin/faro-log
EXPOSE 2025
ENTRYPOINT ["/usr/local/bin/faro-log"]
