"""Hermetic diagnostic checks: no database, HTTP service, Docker or provider access."""
import contextlib
import importlib.util
import io
import os
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import urllib.error

ROOT = Path(__file__).resolve().parents[1]
SPEC = importlib.util.spec_from_file_location('installation_check', ROOT / 'scripts/check-installation.py')
CHECK = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(CHECK)


class Response:
    def __init__(self, code=200, body=b'ok'):
        self.code, self.body = code, body

    def __enter__(self):
        return self

    def __exit__(self, *args):
        pass

    def read(self, limit):
        return self.body[:limit]


class InstallationCheckTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.lines = []
        self.diag = CHECK.Diagnostics(self.root, self.lines.append)

    def fixture(self, path, text='', mode=0o600):
        full = self.root / path
        full.parent.mkdir(parents=True, exist_ok=True)
        full.write_text(text)
        full.chmod(mode)
        return full

    def test_missing_files_and_binary_have_fixes(self):
        with patch.object(CHECK.shutil, 'which', return_value=None):
            self.diag.files()
        self.assertEqual(self.diag.failed, 2)
        self.assertIn('setup.sh --build', '\n'.join(self.lines))
        self.assertEqual(self.diag.finish(), 1)

    def test_old_go_is_rejected_without_building(self):
        self.fixture('go.mod', 'module example\ngo 1.26.0\n')
        with patch.object(CHECK.shutil, 'which', return_value='/fake/go'), patch.object(self.diag, 'command', return_value=(0, 'go version go1.25.0 linux/amd64')) as command:
            self.diag.files()
        self.assertEqual(command.call_args.args[0], ['go', 'version'])
        self.assertEqual(command.call_args.kwargs['env']['GOTOOLCHAIN'], 'local')
        self.assertIn('Go toolchain: Install', '\n'.join(self.lines))

    def test_private_state_after_setup_does_not_require_setup_code(self):
        self.fixture('.data/owner-key.public.pem', 'synthetic-public-key')
        (self.root / '.data').chmod(0o700)
        with patch.dict(os.environ, {'OWNER_KEY_FILE': '.data/owner-key'}):
            self.diag.private_state()
        self.assertEqual(self.diag.failed, 0)
        self.assertNotIn('synthetic-public-key', '\n'.join(self.lines))

    def test_unsafe_key_and_integrity_stop_are_not_repaired(self):
        key = self.fixture('.data/owner-key.public.pem', 'synthetic-public-key', 0o644)
        stop = self.fixture('.data/integrity-stop', 'synthetic-marker')
        with patch.dict(os.environ, {'OWNER_KEY_FILE': '.data/owner-key'}):
            self.diag.private_state()
        self.assertEqual(self.diag.failed, 2)
        self.assertEqual(key.stat().st_mode & 0o777, 0o644)
        self.assertTrue(stop.exists())
        self.assertIn('do not delete', '\n'.join(self.lines))

    def test_database_query_is_read_only_and_url_not_in_arguments(self):
        self.fixture('cmd/server/database_boundary.go', 'const runtimePrivilegeSQL = `SELECT synthetic_catalog_contract`')
        with patch.dict(os.environ, {'DATABASE_URL': 'postgres://synthetic-secret@localhost/db'}), patch.object(CHECK.shutil, 'which', return_value='/fake/psql'), patch.object(self.diag, 'command', return_value=(0, 'f|t|t|t|f|f\n')) as command:
            self.diag.database()
        args, kwargs = command.call_args
        self.assertNotIn('synthetic-secret', str(args))
        self.assertIn('default_transaction_read_only=on', kwargs['env']['PGOPTIONS'])
        self.assertIn('-w', args[0])
        self.assertEqual(self.diag.failed, 0)
        self.assertNotIn('synthetic-secret', '\n'.join(self.lines))

    def test_database_failures_are_redacted_and_bad_privileges_fail(self):
        self.fixture('cmd/server/database_boundary.go', 'const runtimePrivilegeSQL = `SELECT synthetic_catalog_contract`')
        for result in [(1, 'synthetic-sensitive-error'), (0, 't|t|t|t|f|f')]:
            with patch.dict(os.environ, {'DATABASE_URL': 'synthetic-url'}), patch.object(CHECK.shutil, 'which', return_value='/fake/psql'), patch.object(self.diag, 'command', return_value=result):
                self.diag.database()
        self.assertEqual(self.diag.failed, 2)
        self.assertNotIn('synthetic-sensitive-error', '\n'.join(self.lines))

    def test_http_requires_authentication_and_expected_application(self):
        opener = unittest.mock.Mock()
        opener.open.side_effect = [Response(body=b'<title>Agent CommonRoom</title>'), Response(body=b'async function api('), Response(body=b'Help &amp; documentation'), Response(401, b'Sign in required')]
        with patch.object(CHECK.urllib.request, 'build_opener', return_value=opener):
            self.diag.http('http://127.0.0.1:8787')
        self.assertEqual(self.diag.failed, 0)
        self.assertEqual(opener.open.call_count, 4)
        opener.open.side_effect = [Response() for _ in range(4)]
        with patch.object(CHECK.urllib.request, 'build_opener', return_value=opener):
            self.diag.http('http://127.0.0.1:8787')
        self.assertEqual(self.diag.failed, 4)  # wrong pages/assets and unprotected API

    def test_unreachable_http_has_actionable_redacted_errors(self):
        opener = unittest.mock.Mock()
        opener.open.side_effect = urllib.error.URLError('synthetic-private-detail')
        with patch.object(CHECK.urllib.request, 'build_opener', return_value=opener):
            self.diag.http('http://127.0.0.1:8787')
        self.assertEqual(self.diag.failed, 4)
        self.assertIn('SSH tunnel', '\n'.join(self.lines))
        self.assertNotIn('synthetic-private-detail', '\n'.join(self.lines))

    def test_rootful_docker_is_not_accepted(self):
        with patch.object(CHECK.shutil, 'which', return_value='/fake/tool'), patch.object(self.diag, 'command', side_effect=[(0, ''), (0, '["seccomp"]')]):
            self.diag.hybrid()
        self.assertEqual(self.diag.failed, 1)
        self.assertIn('Rootful', '\n'.join(self.lines))

    def test_unsafe_targets_rejected_before_any_checks(self):
        for url in ['http://example.com:8787', 'http://127.0.0.1:8787/path', 'http://user:secret@127.0.0.1:8787', 'https://127.0.0.1:8787', 'http://127.0.0.1']:
            with self.subTest(url=url), patch.object(CHECK, 'Diagnostics') as diag, contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit) as exit_result:
                CHECK.main(['--url', url])
            self.assertEqual(exit_result.exception.code, 2)
            diag.assert_not_called()

    def test_offline_mode_never_contacts_services(self):
        diag = unittest.mock.Mock()
        diag.finish.return_value = 2
        with patch.object(CHECK, 'Diagnostics', return_value=diag):
            self.assertEqual(CHECK.main(['--offline', '--hybrid']), 2)
        diag.files.assert_called_once()
        diag.private_state.assert_called_once()
        diag.database.assert_not_called()
        diag.http.assert_not_called()
        diag.hybrid.assert_not_called()

    def test_command_timeout_does_not_print_backend_details(self):
        with patch.object(CHECK.subprocess, 'run', side_effect=CHECK.subprocess.TimeoutExpired('synthetic-private-command', 12)):
            self.assertEqual(self.diag.command(['synthetic-tool']), (-1, ''))
        self.assertEqual(self.lines, [])

    def test_skipped_checks_are_incomplete_not_success(self):
        self.diag.report('WARN', 'Database', 'Not checked')
        self.assertEqual(self.diag.finish(), 2)
        self.assertIsNone(CHECK.NoRedirect().redirect_request(None, None, 302, '', {}, 'http://example.com'))


if __name__ == '__main__':
    unittest.main()
