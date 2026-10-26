# Build input must be reviewed and digest-pinned; no mutable default or downloads.
ARG GO_IMAGE
FROM ${GO_IMAGE} AS build
ARG GO_IMAGE
RUN case "$GO_IMAGE" in *@sha256:*) ;; *) exit 1 ;; esac; \
    digest="${GO_IMAGE##*@sha256:}"; test "${#digest}" -eq 64; \
    case "$digest" in *[!0-9a-f]*) exit 1 ;; esac
WORKDIR /src
COPY go.mod ./go.mod
COPY cmd/workspace-mount-probe/main.go ./cmd/workspace-mount-probe/main.go
COPY internal/workspacemount/proof.go ./internal/workspacemount/proof.go
COPY internal/workspacemount/observe_linux.go ./internal/workspacemount/observe_linux.go
RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
    go build -trimpath -buildvcs=false -o /out/workspace-mount-probe ./cmd/workspace-mount-probe \
 && mkdir -p /rootfs/workspace && chown 1000:1000 /rootfs/workspace && chmod 0755 /rootfs/workspace
FROM scratch
COPY --from=build /rootfs/ /
COPY --from=build --chmod=0555 /out/workspace-mount-probe /opt/commonroom/workspace-mount-probe
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["/opt/commonroom/workspace-mount-probe"]
