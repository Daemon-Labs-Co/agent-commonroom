#!/usr/bin/env python3
"""Read-only installation diagnostics. Never migrate, log in, launch or repair work."""
import argparse
import ipaddress
import os
from pathlib import Path
import re
import shutil
import stat
import subprocess
import sys
import urllib.error
import urllib.parse
import urllib.request

ROOT = Path(__file__).resolve().parents[1]


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


class Diagnostics:
    def __init__(self, root=ROOT, emit=print):
        self.root = Path(root)
        self.emit = emit
        self.failed = 0
        self.incomplete = 0

    def report(self, level, stage, detail):
        self.emit(f'[{level}] {stage}: {detail}')
        self.failed += level == 'FAIL'
        self.incomplete += level == 'WARN'

    def command(self, args, env=None):
        # Never display command output: tool errors may contain credentials.
        try:
            result = subprocess.run(args, cwd=self.root, env=env, stdin=subprocess.DEVNULL,
                                    capture_output=True, timeout=12, check=False)
            return result.returncode, result.stdout.decode('utf-8', errors='replace')
        except (OSError, subprocess.TimeoutExpired):
            return -1, ''

    def files(self):
        paths = ['go.mod', 'web/index.html', 'web/app.js', 'web/style.css',
                 'web/help.html', 'docs/technical/archive/documentation-index.md', 'migrations/001.sql']
        missing = [p for p in paths if not (self.root / p).is_file()]
        if missing:
            self.report('FAIL', 'Application files', 'Missing ' + ', '.join(missing) +
                        '. Restore the complete checkout; run from its root.')
        else:
            self.report('PASS', 'Application files', 'Required browser, Help and migration assets exist.')
        binary = self.root / 'bin/commonroom'
        if binary.is_file() and os.access(binary, os.X_OK):
            self.report('PASS', 'Controller binary', 'bin/commonroom is executable; binary freshness is not verified.')
        else:
            self.report('FAIL', 'Controller binary', 'Run bash scripts/setup.sh --build. The binary is not launched by this check.')
        if shutil.which('go'):
            rc, output = self.command(['go', 'version'], env=dict(os.environ, GOTOOLCHAIN='local'))
            actual = re.search(r'go(\d+)\.(\d+)', output)
            required = re.search(r'^go (\d+)\.(\d+)', (self.root / 'go.mod').read_text()
                                 if (self.root / 'go.mod').is_file() else '', re.MULTILINE)
            if rc == 0 and actual and required and tuple(map(int, actual.groups())) >= tuple(map(int, required.groups())):
                self.report('PASS', 'Go toolchain', 'Installed version meets the go.mod major/minor requirement.')
            else:
                self.report('FAIL', 'Go toolchain', 'Install the Go version required by go.mod; see docs/technical/archive/installation-options.md.')
        else:
            self.report('WARN', 'Go toolchain', 'Go is missing. A prebuilt controller may run, but setup/build commands need Go.')

    def private_state(self):
        owner = Path(os.environ.get('OWNER_KEY_FILE', '.data/owner-key'))
        if not owner.is_absolute():
            owner = self.root / owner
        if not owner.exists() and not Path(str(owner) + '.public.pem').exists():
            self.report('FAIL', 'Owner setup', 'No setup code/public key found. Start the controller; use its service account and OWNER_KEY_FILE.')
            return
        paths = [owner.parent, owner, Path(str(owner) + '.public.pem'),
                 Path(str(owner) + '.private.pem'), owner.parent / 'human-sessions.json']
        unsafe = False
        for path in paths:
            if path.is_symlink():
                unsafe = True
            elif path.exists():
                metadata = path.stat()
                if metadata.st_mode & 0o077 or metadata.st_uid != os.getuid():
                    unsafe = True
        if unsafe:
            self.report('FAIL', 'Private state permissions', 'Run as the controller account; restrict its private directory to 0700 and key/session files to 0600. Do not replace keys.')
        else:
            self.report('PASS', 'Private state permissions', 'Owner state exists and checked paths are private to this account; key contents were not read.')
        if (owner.parent / 'integrity-stop').exists():
            self.report('FAIL', 'Integrity stop', 'An integrity-stop marker exists. Use Review integrity incident and investigate; do not delete the marker to bypass it.')

    def database(self):
        url = os.environ.get('DATABASE_URL')
        if not url:
            self.report('WARN', 'Database', 'DATABASE_URL is absent in this shell. Run as the service account with its runtime environment; local-config fallback is not inspected.')
            return
        if not shutil.which('psql'):
            self.report('WARN', 'Database', 'Install the PostgreSQL client to check runtime connectivity, schema and privileges.')
            return
        source = self.root / 'cmd/server/database_boundary.go'
        match = re.search(r'const runtimePrivilegeSQL = `([^`]+)`', source.read_text()) if source.is_file() else None
        if not match:
            self.report('FAIL', 'Database checker', 'Restore cmd/server/database_boundary.go from the matching source checkout.')
            return
        env = dict(os.environ, PGDATABASE=url, PGCONNECT_TIMEOUT='5',
                   PGOPTIONS='-c default_transaction_read_only=on -c statement_timeout=5000 -c search_path=pg_catalog,public')
        env.pop('MIGRATION_DATABASE_URL', None)
        rc, output = self.command(['psql', '-X', '-w', '-A', '-t', '-v', 'ON_ERROR_STOP=1',
                                   '-c', match.group(1)], env=env)
        if rc != 0:
            self.report('FAIL', 'Database connection/schema', 'Check runtime URL, private passfile, TLS CA/hostname, network rules and explicit migration. See docs/technical/archive/database-security.md; raw errors are suppressed.')
        elif output.strip() != 'f|t|t|t|f|f':
            self.report('FAIL', 'Database privileges', 'Use the restricted runtime role with SELECT/INSERT and sequence USAGE; remove ownership, memberships and mutation/DDL rights. Apply documented grants after migration.')
        else:
            self.report('PASS', 'Database', 'Read-only catalog query passed the controller runtime privilege contract; no records were read or written.')

    def http(self, base):
        # No ambient proxy, cookies, authentication, redirects or non-loopback targets.
        opener = urllib.request.build_opener(urllib.request.ProxyHandler({}), NoRedirect())
        checks = [('/', 200, b'Agent CommonRoom', 'Browser page'),
                  ('/app.js', 200, b'async function api(', 'Browser script'),
                  ('/help/', 200, b'Help &amp; documentation', 'Help page'),
                  ('/api/state', 401, b'Sign in required', 'Authentication boundary')]
        for path, expected, marker, stage in checks:
            try:
                try:
                    response = opener.open(base + path, timeout=5)
                except urllib.error.HTTPError as error:
                    response = error
                with response:
                    status_code = response.code
                    body = response.read(65536)
                if status_code == expected and (marker is None or marker in body) and (marker is not None or body):
                    self.report('PASS', stage, 'Expected unauthenticated HTTP response received.')
                else:
                    self.report('FAIL', stage, f'Unexpected HTTP {status_code}. Check the matching controller, asset paths and nginx Host/Origin forwarding. Never bypass authentication.')
            except (OSError, ValueError):
                self.report('FAIL', stage, 'Cannot reach the loopback service. Start the controller; check its startup logs, LISTEN_ADDR, port and SSH tunnel.')

    def hybrid(self):
        for tool in ('nginx', 'docker', 'ssh'):
            if not shutil.which(tool):
                self.report('FAIL', 'Hybrid ' + tool, 'Install/configure this tool on the Linux host; see docs/technical/archive/quick-start.md.')
        if shutil.which('nginx'):
            rc, _ = self.command(['nginx', '-t'])
            self.report('PASS' if rc == 0 else 'FAIL', 'nginx configuration',
                        'Configuration syntax check passed.' if rc == 0 else 'Run nginx -t under the authorized account to inspect syntax/access errors. No reload was attempted.')
        if shutil.which('docker'):
            rc, output = self.command(['docker', 'info', '--format', '{{json .SecurityOptions}}'])
            self.report('PASS' if rc == 0 and 'rootless' in output else 'FAIL', 'Docker identity',
                        'Current Docker context reports rootless mode; resource limits and worker admission still need verification.'
                        if rc == 0 and 'rootless' in output else 'Use the approved worker account/rootless Docker context. Rootful or unreachable Docker does not satisfy agent isolation.')

    def finish(self):
        self.emit('Manual acceptance: sign in, open a room, and use Connections > Test connection for each approved worker/provider. Login, audit integrity, remote resource limits and live provider execution are not certified by this script.')
        if self.failed:
            self.emit(f'Result: {self.failed} failed check(s), {self.incomplete} incomplete check(s). Fix the reported stages and rerun.')
            return 1
        if self.incomplete:
            self.emit(f'Result: checked stages passed, but {self.incomplete} stage(s) remain unverified.')
            return 2
        self.emit('Result: all automated installation checks passed. Complete manual acceptance before relying on the installation.')
        return 0


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--url', default='http://127.0.0.1:8787', help='Numeric-loopback HTTP URL of the controller or private tunnel')
    parser.add_argument('--hybrid', action='store_true', help='Also check local nginx syntax and current Docker rootless identity')
    parser.add_argument('--offline', action='store_true', help='Check local files/state only; result remains incomplete')
    args = parser.parse_args(argv)
    try:
        parsed = urllib.parse.urlsplit(args.url)
        if (parsed.scheme != 'http' or not ipaddress.ip_address(parsed.hostname).is_loopback
                or parsed.username or parsed.password or parsed.path not in ('', '/')
                or parsed.query or parsed.fragment or parsed.port is None):
            raise ValueError()
    except (TypeError, ValueError):
        parser.error('--url must be an HTTP numeric-loopback address with an explicit port and no credentials/path/query')
    if os.name != 'posix':
        parser.error('Run inside Linux/WSL2 or macOS, using the controller service account')
    diag = Diagnostics()
    try:
        diag.files()
        diag.private_state()
        if args.offline:
            diag.report('WARN', 'Online checks', 'Database and HTTP checks were skipped by --offline.')
        else:
            diag.database()
            diag.http(args.url.rstrip('/'))
        if args.hybrid:
            if args.offline:
                diag.report('WARN', 'Hybrid services', 'Service commands skipped by --offline.')
            else:
                diag.hybrid()
    except OSError:
        diag.report('FAIL', 'Local access', 'A required file cannot be read or inspected. Use the service account and restore the matching checkout; private paths are suppressed.')
    return diag.finish()


if __name__ == '__main__':
    sys.exit(main())
