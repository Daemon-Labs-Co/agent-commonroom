# Digest-pinned reviewed compiler input; no dependency downloads or providers.
ARG GO_IMAGE
FROM ${GO_IMAGE} AS build
ARG GO_IMAGE
RUN case "$GO_IMAGE" in *@sha256:*) ;; *) exit 1 ;; esac; \
    digest="${GO_IMAGE##*@sha256:}"; test "${#digest}" -eq 64; \
    case "$digest" in *[!0-9a-f]*) exit 1 ;; esac
WORKDIR /src
COPY go.mod ./go.mod
COPY cmd/object-read/main.go ./cmd/object-read/main.go
COPY internal/objectread/reader.go ./internal/objectread/reader.go
COPY internal/objectcustody/types.go internal/objectcustody/store_unix.go internal/objectcustody/read_channel.go internal/objectcustody/read_journal_channel.go ./internal/objectcustody/
COPY internal/workspacemount/proof.go internal/workspacemount/observe_linux.go ./internal/workspacemount/
RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
    go build -trimpath -buildvcs=false -o /out/object-read ./cmd/object-read \
 && mkdir -p /rootfs/objects \
 && chown 1000:1000 /rootfs/objects && chmod 0700 /rootfs/objects
FROM scratch
COPY LICENSE /licenses/Agent-CommonRoom-LICENSE
COPY --from=build /usr/local/go/LICENSE /licenses/Go-LICENSE
COPY --from=build /rootfs/ /
COPY --from=build --chmod=0555 /out/object-read /opt/commonroom/object-read
USER 1000:1000
WORKDIR /objects
ENTRYPOINT ["/opt/commonroom/object-read"]
