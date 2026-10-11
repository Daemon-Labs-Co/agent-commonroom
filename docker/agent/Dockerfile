# Pin the base digest and both provider versions when building a trusted image.
ARG NODE_IMAGE
FROM ${NODE_IMAGE}
ARG CODEX_VERSION
ARG CLAUDE_VERSION
RUN test -n "$CODEX_VERSION" && test -n "$CLAUDE_VERSION" \
 && npm install --global --allow-scripts=@anthropic-ai/claude-code "@openai/codex@$CODEX_VERSION" "@anthropic-ai/claude-code@$CLAUDE_VERSION" \
 && mkdir -p /home/agent /workspace && chown 1000:1000 /home/agent /workspace
COPY docker/agent/entrypoint.cjs /opt/commonroom/entrypoint.cjs
USER 1000:1000
WORKDIR /workspace
ENTRYPOINT ["node", "/opt/commonroom/entrypoint.cjs"]
