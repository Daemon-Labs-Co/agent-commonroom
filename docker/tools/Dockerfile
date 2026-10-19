# Reviewed immutable base images are mandatory; no mutable-tag defaults.
# This recipe is not wired to production. Do not reuse the provider/auth image.
ARG GO_IMAGE
ARG NODE_IMAGE
FROM ${GO_IMAGE} AS build
ARG GO_IMAGE
ARG NODE_IMAGE
RUN for image in "$GO_IMAGE" "$NODE_IMAGE"; do \
      case "$image" in *@sha256:*) ;; *) exit 1 ;; esac; \
      digest="${image##*@sha256:}"; \
      test "${#digest}" -eq 64 || exit 1; \
      case "$digest" in *[!0-9a-f]*) exit 1 ;; esac; \
    done
WORKDIR /src
COPY go.mod ./go.mod
COPY cmd/tool-helper/main.go ./cmd/tool-helper/main.go
COPY cmd/workspace-bootstrap/main.go ./cmd/workspace-bootstrap/main.go
COPY internal/workspacebootstrap/bootstrap.go ./internal/workspacebootstrap/bootstrap.go
COPY internal/toolworkspace/workspace_unix.go ./internal/toolworkspace/workspace_unix.go
COPY internal/strictjson/encoding.go ./internal/strictjson/encoding.go
# The helper imports only the standard library and toolworkspace. No module,
# toolchain, provider or package downloads are allowed during this build step.
RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
    go build -trimpath -buildvcs=false -o /out/tool-helper ./cmd/tool-helper \
 && CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off \
    go build -trimpath -buildvcs=false -o /out/workspace-bootstrap ./cmd/workspace-bootstrap

FROM ${NODE_IMAGE}
# The reviewed Linux Node base must provide this exact path and no baked secrets.
# Pinning alone does not establish provenance or absence of provider credentials.
RUN test -x /usr/local/bin/node \
 && test -x /bin/mkdir \
 && test -x /usr/bin/git \
 && mkdir -p /opt/commonroom /workspace \
 && chown 1000:1000 /workspace \
 && chmod 0755 /opt/commonroom /workspace
COPY --from=build /usr/local/go/ /usr/local/go/
COPY --from=build --chmod=0555 /out/tool-helper /opt/commonroom/tool-helper
COPY --from=build --chmod=0555 /out/workspace-bootstrap /opt/commonroom/workspace-bootstrap
COPY --chmod=0444 docker/tools/tool-call.cjs /opt/commonroom/tool-call.cjs
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/node", "/opt/commonroom/tool-call.cjs"]
