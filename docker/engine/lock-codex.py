"""Patch only the reviewed Codex release for a bridge-only engine build.

This build-time tool accepts trusted source paths, never a model request. Source
changes CAN restore native handlers BUT invalidate credential separation. There
is no opt-out or runtime toggle. The supervisor still owns scopes/audit/isolation.
"""
import hashlib
from pathlib import Path
import sys

UPSTREAM_COMMIT = "092d3acd6bec3e3a14bdc7e7a2810ab628ab759d"
SOURCE_FILE = "codex-rs/core/src/tools/spec_plan.rs"
SOURCE_SHA256 = "db040c0c07160af64bebd5a7a5071658cb7ff58db341edd70e2795958b1d07f3"
MARKER = "// COMMONROOM_BRIDGE_ONLY_V1"


def transform(original: bytes, fragment: str) -> str:
    """Verify upstream bytes before changing the sole production tool factory."""
    if hashlib.sha256(original).hexdigest() != SOURCE_SHA256:
        raise ValueError("Codex source does not match reviewed release")
    text = original.decode("utf-8")
    begin = text.index("pub(crate) fn build_tool_router(")
    body = text.index(") -> CodexResult<ToolRouter> {", begin) + len(") -> CodexResult<ToolRouter> {")
    end = text.index("\n}\n", body)
    replacement = "\n    " + MARKER + "\n    commonroom_bridge_router(&turn_context.dynamic_tools, Arc::clone(&session.tool_policy))"
    return text[:body] + replacement + text[end:] + "\n" + fragment


def apply(source: Path, fragment_path: Path) -> None:
    target = source / SOURCE_FILE
    if target.is_symlink() or not target.is_file():
        raise ValueError("Reviewed source file unavailable")
    fragment = fragment_path.read_text()
    target.write_text(transform(target.read_bytes(), fragment))


if __name__ == "__main__":
    try:
        if len(sys.argv) != 2:
            raise ValueError("Trusted source checkout required")
        apply(Path(sys.argv[1]), Path(__file__).with_name("codex-bridge-router.rs"))
        print("Reviewed Codex bridge-only source prepared; compile and verify before use")
    except (OSError, UnicodeError, ValueError):
        print("Codex bridge-only source preparation refused", file=sys.stderr)
        sys.exit(2)
