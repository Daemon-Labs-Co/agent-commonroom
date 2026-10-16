"""Exercise setup helpers with fake tools; no package, DB, Docker or provider activity."""
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]


class SetupScriptsTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        shutil.copytree(ROOT / 'scripts', self.root / 'scripts')
        self.tools = self.root / 'fake-tools'
        self.tools.mkdir()
        self.log = self.root / 'calls'
        self.env = dict(os.environ, PATH=str(self.tools) + os.pathsep + os.environ['PATH'],
                        SETUP_CALLS=str(self.log))
        self.env.pop('DATABASE_URL', None)
        self.env.pop('MIGRATION_DATABASE_URL', None)
        self.tool('uname', '#!/bin/sh\necho "${SETUP_OS:-Linux}"\n')
        self.tool('go', '''#!/bin/sh
printf '%s\n' "$*" >> "$SETUP_CALLS"
if [ "$1" = version ]; then echo 'go version go1.26.0 test/test'; exit 0; fi
if [ "$1" = build ]; then
  [ "${SETUP_FAIL_BUILD:-}" != 1 ] || exit 9
  cat > bin/commonroom <<'RUN'
#!/bin/sh
printf 'binary:%s migration-env:%s\n' "$*" "${MIGRATION_DATABASE_URL:+present}" >> "$SETUP_CALLS"
RUN
  chmod +x bin/commonroom
fi
''')

    def tool(self, name, source):
        p = self.tools / name
        p.write_text(source)
        p.chmod(0o700)

    def run_setup(self, script='setup.sh', *args):
        return subprocess.run(['bash', str(self.root / 'scripts' / script), *args],
                              cwd=self.temp.name, env=self.env, text=True,
                              capture_output=True, timeout=5)

    def test_check_does_not_build_or_run(self):
        result = self.run_setup('setup.sh', '--check')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertEqual(self.log.read_text().strip(), 'version')
        self.assertFalse((self.root / 'bin').exists())

    def test_default_build_resolves_checkout_root(self):
        result = self.run_setup()
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('build -o bin/commonroom ./cmd/server', self.log.read_text())
        self.assertNotIn('binary:', self.log.read_text())

    def test_missing_urls_do_not_build_migrate_or_start(self):
        for mode in ('--migrate', '--start'):
            with self.subTest(mode=mode):
                self.log.unlink(missing_ok=True)
                result = self.run_setup('setup.sh', mode)
                self.assertNotEqual(result.returncode, 0)
                self.assertNotIn('build', self.log.read_text())

    def test_migrate_is_explicit_and_start_removes_owner_url(self):
        self.env['MIGRATION_DATABASE_URL'] = 'synthetic-owner-url'
        result = self.run_setup('setup.sh', '--migrate')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('binary:--migrate migration-env:present', self.log.read_text())
        self.env['DATABASE_URL'] = 'synthetic-runtime-url'
        result = self.run_setup('setup.sh', '--start')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn('binary: migration-env:\n', self.log.read_text())
        self.assertNotIn('synthetic-owner-url', result.stdout + result.stderr)

    def test_invalid_mode_and_build_failure_never_launch(self):
        result = self.run_setup('setup.sh', '--invalid')
        self.assertEqual(result.returncode, 2)
        self.assertFalse(self.log.exists())
        self.env['DATABASE_URL'] = 'synthetic-runtime-url'
        self.env['SETUP_FAIL_BUILD'] = '1'
        result = self.run_setup('setup.sh', '--start')
        self.assertEqual(result.returncode, 9)
        self.assertNotIn('binary:', self.log.read_text())

    def test_platform_wrappers(self):
        for script, platform in [('setup-macos.sh', 'Darwin'), ('setup-linux.sh', 'Linux')]:
            self.env['SETUP_OS'] = platform
            result = self.run_setup(script, '--check')
            self.assertEqual(result.returncode, 0, result.stderr)
            self.env['SETUP_OS'] = 'Other'
            self.assertNotEqual(self.run_setup(script, '--check').returncode, 0)

    def test_hybrid_checks_tools_without_reconfiguring(self):
        for name in ('nginx', 'docker', 'psql', 'ssh'):
            self.tool(name, '#!/bin/sh\necho unexpected-execution >&2\nexit 99\n')
        result = self.run_setup('setup-hybrid.sh', '--check')
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertNotIn('unexpected-execution', result.stderr)
        self.assertEqual(self.log.read_text().strip(), 'version')


if __name__ == '__main__':
    unittest.main()
