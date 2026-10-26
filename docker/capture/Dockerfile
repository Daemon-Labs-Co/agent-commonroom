# Reviewed digest-pinned compiler input only. No dependency downloads or providers.
ARG GO_IMAGE
FROM ${GO_IMAGE} AS build
ARG GO_IMAGE
RUN case "$GO_IMAGE" in *@sha256:*) ;; *) exit 1 ;; esac; \
    digest="${GO_IMAGE##*@sha256:}"; test "${#digest}" -eq 64; \
    case "$digest" in *[!0-9a-f]*) exit 1 ;; esac
WORKDIR /src
COPY go.mod ./go.mod
COPY cmd/workspace-capture/main.go ./cmd/workspace-capture/main.go
COPY internal/workspacecapture/frozen.go internal/workspacecapture/custody.go internal/workspacecapture/channel.go ./internal/workspacecapture/
COPY internal/workspacesnapshot/snapshot_unix.go ./internal/workspacesnapshot/snapshot_unix.go
COPY internal/objectcustody/types.go internal/objectcustody/store_unix.go ./internal/objectcustody/
COPY internal/workspacemount/proof.go internal/workspacemount/observe_linux.go ./internal/workspacemount/
RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
    go build -trimpath -buildvcs=false -o /out/workspace-capture ./cmd/workspace-capture \
 && mkdir -p /rootfs/workspace /rootfs/objects \
 && chown 1000:1000 /rootfs/workspace /rootfs/objects \
 && chmod 0755 /rootfs/workspace && chmod 0700 /rootfs/objects
FROM scratch
COPY LICENSE /licenses/Agent-CommonRoom-LICENSE
COPY --from=build /usr/local/go/LICENSE /licenses/Go-LICENSE
COPY --from=build /rootfs/ /
COPY --from=build --chmod=0555 /out/workspace-capture /opt/commonroom/workspace-capture
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["/opt/commonroom/workspace-capture"]
