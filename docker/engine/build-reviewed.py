"""Build the reviewed bridge-only Codex source without provider execution.

Operator build infrastructure only: the source checkout is trusted public input.
Never point this at a live repository/home or mount credentials/Docker sockets.
Changing this launcher CAN remove confinement BUT invalidates build provenance.
"""
import argparse
import importlib.util
import json
import os
from pathlib import Path
import re
import subprocess
import sys

RUST_IMAGE = 'rust@sha256:4c2fd73ef19c5ef9d54bee03b06b2839a392604fbfcd578ed948b71b37c1d7fb'
BUILD_COMMAND = ('cargo test --lib -p codex-core commonroom_bridge_tests --locked && '
                 'cargo build -p codex-cli --bin codex --release --locked')


def build_args(source, name):
    source = Path(source)
    if (not source.is_absolute() or source.is_symlink() or not source.is_dir()
            or source.resolve() != source or source in (Path('/'), Path.home())
            or any(c in str(source) for c in ',\r\n')
            or not re.fullmatch(r'commonroom-engine-build-[a-z0-9-]{1,40}', name)):
        raise ValueError('Dedicated absolute build source and unique build name required')
    # Refuse common accidental secret/checkouts at the mount root. This is not a
    # secret scanner or a substitute for reviewing the complete source tree.
    if any((source / entry).exists() for entry in ('.git', '.codex', '.ssh', '.aws', '.env')):
        raise ValueError('Use a fresh public source export without private configuration')
    for relative in ('codex-rs', 'codex-rs/core', 'codex-rs/core/src', 'codex-rs/core/src/tools'):
        if (source / relative).is_symlink():
            raise ValueError('Source preparation paths must remain inside the export')
    if not (source / 'codex-rs/Cargo.toml').is_file():
        raise ValueError('Codex public source export required')
    return ['docker', 'run', '--name', name, '--cpus=4', '--memory=8g',
            '--memory-swap=8g', '--pids-limit=512', '--cap-drop=ALL', '--security-opt=no-new-privileges',
            '--mount', f'type=bind,src={source},dst=/src', '--workdir', '/src/codex-rs',
            '--env', 'CARGO_HOME=/build-cache/cargo',
            '--env', 'CARGO_TARGET_DIR=/build-cache/target',
            '--env', 'CARGO_BUILD_JOBS=1',
            '--tmpfs', '/tmp:rw,nosuid,nodev,size=1g', RUST_IMAGE,
            'sh', '-c', BUILD_COMMAND]


def run(source, name, prepare, invoke=subprocess.run):
    args = build_args(source, name)
    # Check actual daemon mode before writing source or creating a container.
    env = {key: os.environ[key] for key in ('PATH', 'HOME', 'XDG_RUNTIME_DIR') if key in os.environ}
    result = invoke(['docker', 'info', '--format', '{{json .}}'],
                    check=True, capture_output=True, text=True, env=env)
    info = json.loads(result.stdout)
    if not isinstance(info, dict):
        raise ValueError('Docker capability evidence required')
    options = info.get('SecurityOptions')
    if not isinstance(options, list) or 'name=rootless' not in options:
        raise ValueError('Rootless Docker is required; no host build fallback')
    if any(info.get(key) is not True for key in ('MemoryLimit', 'SwapLimit', 'PidsLimit', 'CpuCfsQuota')):
        raise ValueError('Enforceable memory, swap, PID and CPU limits are required')
    prepare(Path(source))
    # Use a reduced process environment: no inherited provider keys or custom
    # Docker endpoint/context. Docker uses its default rootless local socket.
    invoke(args, check=True, env=env)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--source', required=True, type=Path)
    parser.add_argument('--name', required=True)
    args = parser.parse_args()
    spec = importlib.util.spec_from_file_location('codex_source_lock', Path(__file__).with_name('lock-codex.py'))
    lock = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(lock)
    try:
        run(args.source, args.name,
            lambda source: lock.apply(source, Path(__file__).with_name('codex-bridge-router.rs')))
    except (OSError, ValueError, subprocess.SubprocessError):
        print('Reviewed engine build failed; retain diagnostics and do not deploy the artifact', file=sys.stderr)
        return 1
    print('Build/tests completed; image provenance and live acceptance are still required')
    return 0


if __name__ == '__main__':
    sys.exit(main())
