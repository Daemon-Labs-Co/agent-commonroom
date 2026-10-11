# Pin the base digest and both provider versions when building a trusted image.
ARG NODE_IMAGE
FROM ${NODE_IMAGE}
ARG CODEX_VERSION
ARG CLAUDE_VERSION
# Native provider clients need the system trust store; Node's bundled roots
# alone do not make device login work. Never disable TLS verification.
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates \
 && rm -rf /var/lib/apt/lists/* && test -s /etc/ssl/certs/ca-certificates.crt
RUN test -n "$CODEX_VERSION" && test -n "$CLAUDE_VERSION" \
 && npm install --global --allow-scripts=@anthropic-ai/claude-code "@openai/codex@$CODEX_VERSION" "@anthropic-ai/claude-code@$CLAUDE_VERSION" \
 && mkdir -p /home/agent /workspace && chown 1000:1000 /home/agent /workspace
COPY docker/agent/entrypoint.cjs /opt/commonroom/entrypoint.cjs
# Normalize both directory traversal and file readability regardless of the
# build context umask; COPY --chmod also applies to newly created parents.
RUN chmod 0755 /opt/commonroom && chmod 0644 /opt/commonroom/entrypoint.cjs
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["node", "/opt/commonroom/entrypoint.cjs"]
