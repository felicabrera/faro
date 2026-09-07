# FARO transparency log image.
#
# Multi-stage, ending at distroless/static as a non-root user: the runtime layer
# has no shell, no package manager and no libc.
#
# The image ships only the log service. The audit explorer in web/ is a static
# Next.js app and is deployed separately, behind a CDN; keeping them apart means
# a compromise of the public website cannot reach the log's signing key.

FROM golang:1.27-bookworm AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/faro-log ./cmd/faro-log

# The final stage has no shell to create FARO_STORAGE_DIR at startup, and a
# named Docker volume mounts as root-owned by default, which the nonroot user
# cannot write to. Pre-create it here, owned by distroless's nonroot
# uid/gid (65532), so a fresh volume is writable on first run.
RUN mkdir -p /out/data && chown 65532:65532 /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/faro-log /usr/local/bin/faro-log
COPY --from=build --chown=65532:65532 /out/data /data
EXPOSE 2025
ENTRYPOINT ["/usr/local/bin/faro-log"]
