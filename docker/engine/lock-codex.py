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
FRAGMENT_SHA256 = "3536a2340d9359b2450760b3078405ee0254f201094938e5a65ccc56fa3009d6"
LOCK_SHA256 = "49fe417bcafe3f69f1f6636e23274f90d37e75e388127a76888c39001a4f8117"
NORMALIZED_LOCK_SHA256 = "765110795bf3385d734bf1acfcae77660d29a6c678d5b77015d7ffe49f26e69f"
MARKER = "// COMMONROOM_BRIDGE_ONLY_V1"


def transform(original: bytes, fragment: str) -> str:
    """Verify upstream bytes before changing the sole production tool factory."""
    if hashlib.sha256(original).hexdigest() != SOURCE_SHA256:
        raise ValueError("Codex source does not match reviewed release")
    if hashlib.sha256(fragment.encode("utf-8")).hexdigest() != FRAGMENT_SHA256:
        raise ValueError("Bridge policy does not match reviewed source")
    text = original.decode("utf-8")
    begin = text.index("pub(crate) fn build_tool_router(")
    body = text.index(") -> CodexResult<ToolRouter> {", begin) + len(") -> CodexResult<ToolRouter> {")
    end = text.index("\n}\n", body)
    replacement = "\n    " + MARKER + "\n    commonroom_bridge_router(&turn_context.dynamic_tools, Arc::clone(&session.tool_policy))"
    return text[:body] + replacement + text[end:] + "\n" + fragment


def normalize_lock(original: bytes) -> bytes:
    """Normalize release-local versions without resolving/updating dependencies."""
    if hashlib.sha256(original).hexdigest() != LOCK_SHA256:
        raise ValueError("Lockfile does not match reviewed release")
    chunks = original.split(b"[[package]]")
    for index, chunk in enumerate(chunks):
        if index and b"\nsource = " not in chunk:
            chunks[index] = chunk.replace(b'\nversion = "0.0.0"\n', b'\nversion = "0.162.1"\n', 1)
    result = b"[[package]]".join(chunks)
    if hashlib.sha256(result).hexdigest() != NORMALIZED_LOCK_SHA256:
        raise ValueError("Unexpected release lock normalization")
    return result


def apply(source: Path, fragment_path: Path) -> None:
    target = source / SOURCE_FILE
    if target.is_symlink() or not target.is_file():
        raise ValueError("Reviewed source file unavailable")
    if fragment_path.is_symlink() or not fragment_path.is_file():
        raise ValueError("Reviewed bridge policy unavailable")
    fragment = fragment_path.read_bytes().decode("utf-8")
    result = transform(target.read_bytes(), fragment)
    lockfile = source / "codex-rs/Cargo.lock"
    if lockfile.is_symlink() or not lockfile.is_file():
        raise ValueError("Reviewed lockfile unavailable")
    normalized = normalize_lock(lockfile.read_bytes())
    # Validate every input before writing either file. Interrupted writes require
    # a fresh checkout; preparation is never treated as a deployable artifact.
    target.write_bytes(result.encode("utf-8"))
    lockfile.write_bytes(normalized)


if __name__ == "__main__":
    try:
        if len(sys.argv) != 2:
            raise ValueError("Trusted source checkout required")
        apply(Path(sys.argv[1]), Path(__file__).with_name("codex-bridge-router.rs"))
        print("Reviewed Codex bridge-only source prepared; compile and verify before use")
    except (OSError, UnicodeError, ValueError):
        print("Codex bridge-only source preparation refused", file=sys.stderr)
        sys.exit(2)
