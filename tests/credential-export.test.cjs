const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
// Exercise the exact fixed worker script without Docker, providers or a DB.
const source = fs.readFileSync(path.join(__dirname, '../cmd/server/worker_auth.go'), 'utf8');
const script = source.match(/script := `([^`]+)`/)[1];
for (const kind of ['valid', 'missing', 'symlink', 'oversized', 'invalid', 'directory']) {
 test(`private credential export: ${kind}`, () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'commonroom-export-test-'));
  try {
   const file = path.join(dir, 'credential.json');
   const fixture = JSON.stringify({ synthetic: true });
   if (kind === 'valid') fs.writeFileSync(file, fixture, { mode: 0o600 });
   if (kind === 'invalid') fs.writeFileSync(file, 'not-json');
   if (kind === 'oversized') fs.writeFileSync(file, 'x'.repeat(65537));
   if (kind === 'directory') fs.mkdirSync(file);
   if (kind === 'symlink') {
    fs.writeFileSync(path.join(dir, 'target'), fixture);
    fs.symlinkSync(path.join(dir, 'target'), file);
   }
   const result = spawnSync(process.execPath, ['-e', script, file], { encoding: 'utf8', timeout: 3000 });
   assert.equal(result.status, kind === 'valid' ? 0 : 78);
   assert.equal(result.stdout, kind === 'valid' ? fixture : '');
   assert.equal(result.stderr, '', 'no credential or file diagnostics may escape');
  } finally { fs.rmSync(dir, { recursive: true, force: true }); }
 });
}
