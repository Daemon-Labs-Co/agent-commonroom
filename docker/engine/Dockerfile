# Trusted build inputs only. This is not a stock CLI image or an auto-upgrader.
# ENGINE_SOURCE_IMAGE must be the reviewed bridge-only source build, publishing
# /out/codex and its upstream /out/LICENSE. No provider credentials in either base.
ARG ENGINE_SOURCE_IMAGE
ARG NODE_IMAGE
FROM ${ENGINE_SOURCE_IMAGE} AS reviewed_engine
FROM ${NODE_IMAGE}
ARG ENGINE_SOURCE_IMAGE
ARG NODE_IMAGE
ARG CODEX_BINARY_SHA256
RUN for image in "$ENGINE_SOURCE_IMAGE" "$NODE_IMAGE"; do \
      case "$image" in *@sha256:*) ;; *) exit 1 ;; esac; \
      digest="${image##*@sha256:}"; test "${#digest}" -eq 64 || exit 1; \
      case "$digest" in *[!0-9a-f]*) exit 1 ;; esac; \
    done \
 && test "${#CODEX_BINARY_SHA256}" -eq 64 \
 && case "$CODEX_BINARY_SHA256" in *[!0-9a-f]*) exit 1 ;; esac \
 && test -x /usr/local/bin/node \
 && test ! -e /etc/codex && test ! -e /.codex && test ! -e /AGENTS.md \
 && mkdir -p /opt/commonroom/bin /opt/commonroom/notices \
 && chmod 0755 /opt/commonroom /opt/commonroom/bin /opt/commonroom/notices
COPY --from=reviewed_engine --chmod=0555 /out/codex /opt/commonroom/bin/codex
COPY --from=reviewed_engine --chmod=0444 /out/LICENSE /opt/commonroom/notices/codex-LICENSE
COPY --from=reviewed_engine /out/notices/ /opt/commonroom/notices/upstream/
RUN printf '%s  %s\n' "$CODEX_BINARY_SHA256" /opt/commonroom/bin/codex | sha256sum --check --strict -
COPY --chmod=0444 docker/engine/entrypoint.cjs /opt/commonroom/engine-entrypoint.cjs
USER 1000:1000
WORKDIR /tmp
ENTRYPOINT ["/usr/local/bin/node", "/opt/commonroom/engine-entrypoint.cjs"]

CMD []
