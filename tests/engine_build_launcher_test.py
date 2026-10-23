import importlib.util
import json
import os
from pathlib import Path
import subprocess
import tempfile
import types
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location('build_reviewed', Path(__file__).resolve().parents[1] / 'docker/engine/build-reviewed.py')
launcher = importlib.util.module_from_spec(spec)
spec.loader.exec_module(launcher)


class EngineBuildLauncherTests(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.addCleanup(self.tmp.cleanup)
        self.source = Path(self.tmp.name).resolve() / 'source'
        (self.source / 'codex-rs').mkdir(parents=True)
        (self.source / 'codex-rs/Cargo.toml').write_text('[workspace]')

    def test_build_is_confined_and_credentials_are_not_inherited(self):
        calls = []
        prepared = []
        def invoke(args, **kwargs):
            calls.append((args, kwargs))
            return types.SimpleNamespace(stdout=json.dumps(['name=rootless']))
        with patch.dict(os.environ, {'OPENAI_API_KEY': 'synthetic-canary', 'DOCKER_HOST': 'tcp://foreign.invalid:2375'}):
            launcher.run(self.source, 'commonroom-engine-build-test', prepared.append, invoke)
        self.assertEqual(prepared, [self.source])
        self.assertEqual(len(calls), 2)
        args, kwargs = calls[1]
        self.assertEqual(args.count('--mount'), 1)
        self.assertIn(f'type=bind,src={self.source},dst=/src', args)
        for value in ('--cpus=4', '--memory=8g', '--memory-swap=8g', 'CARGO_BUILD_JOBS=4', '--pids-limit=512', '--cap-drop=ALL', '--security-opt=no-new-privileges', launcher.RUST_IMAGE):
            self.assertIn(value, args)
        self.assertEqual(args[-1], launcher.BUILD_COMMAND)
        for _, options in calls:
            self.assertNotIn('OPENAI_API_KEY', options['env'])
            self.assertNotIn('DOCKER_HOST', options['env'])
        self.assertNotIn('--privileged', args)
        self.assertNotIn('/var/run/docker.sock', ' '.join(args))

    def test_non_rootless_and_malformed_evidence_stop_before_preparation(self):
        for output in ('["name=seccomp"]', '{}', 'invalid'):
            calls = []
            def invoke(args, **kwargs):
                calls.append(args)
                return types.SimpleNamespace(stdout=output)
            with self.subTest(output=output), self.assertRaises(ValueError):
                launcher.run(self.source, 'commonroom-engine-build-test', lambda _: self.fail('source changed'), invoke)
            self.assertEqual(len(calls), 1)

    def test_preparation_failure_never_starts_container(self):
        calls = []
        def invoke(args, **kwargs):
            calls.append(args)
            return types.SimpleNamespace(stdout='["name=rootless"]')
        def prepare(_):
            raise ValueError('pin mismatch')
        with self.assertRaises(ValueError):
            launcher.run(self.source, 'commonroom-engine-build-test', prepare, invoke)
        self.assertEqual(len(calls), 1)

    def test_unsafe_sources_and_mount_spellings_refused(self):
        for entry in ('.git', '.codex', '.ssh', '.aws', '.env'):
            marker = self.source / entry
            marker.touch()
            with self.subTest(entry=entry), self.assertRaises(ValueError):
                launcher.build_args(self.source, 'commonroom-engine-build-test')
            marker.unlink()
        alias = Path(self.tmp.name) / 'alias'
        alias.symlink_to(self.source, target_is_directory=True)
        for source in (alias, Path('relative'), Path('/'), self.source / '..' / 'source'):
            with self.subTest(source=source), self.assertRaises(ValueError):
                launcher.build_args(source, 'commonroom-engine-build-test')
        with self.assertRaises(ValueError):
            launcher.build_args(self.source, '--privileged')

    def test_preparation_parent_symlink_is_refused(self):
        external = Path(self.tmp.name) / 'external'
        external.mkdir()
        (self.source / 'codex-rs/core').symlink_to(external, target_is_directory=True)
        with self.assertRaises(ValueError):
            launcher.build_args(self.source, 'commonroom-engine-build-test')

    def test_failed_compilation_is_not_success(self):
        def invoke(args, **kwargs):
            if args[1] == 'run':
                raise subprocess.CalledProcessError(1, args)
            return types.SimpleNamespace(stdout='["name=rootless"]')
        with self.assertRaises(subprocess.CalledProcessError):
            launcher.run(self.source, 'commonroom-engine-build-test', lambda _: None, invoke)
