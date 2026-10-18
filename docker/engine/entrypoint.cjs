'use strict';

// Reviewed credential-bearing engine bootstrap, never a host tool API. The
// worker must prove this runs in its fresh isolated engine container without a
// repository/tool/audit/Docker mount. Source changes CAN bypass this policy,
// BUT invalidate mandatory credential separation and reviewed-image admission.
const fs = require('node:fs');
const net = require('node:net');
const cp = require('node:child_process');
const HOME = '/tmp/commonroom-engine';
const CODEX_HOME = `${HOME}/.codex`;

function environment() {
  return {
    HOME, CODEX_HOME, PATH: '/opt/commonroom/bin:/usr/local/bin:/usr/bin:/bin',
    HTTP_PROXY: 'http://127.0.0.1:18080', HTTPS_PROXY: 'http://127.0.0.1:18080',
    ALL_PROXY: '', NO_PROXY: '', NODE_USE_ENV_PROXY: '1',
    TMPDIR: '/tmp', XDG_CONFIG_HOME: `${HOME}/.config`,
    XDG_CACHE_HOME: `${HOME}/.cache`, XDG_DATA_HOME: `${HOME}/.local/share`
  };
}
function invocation() {
  return ['--disable', 'respect_system_proxy', '--disable', 'system_proxy_fallback',
    '--disable', 'hooks', '--disable', 'plugins', '--disable', 'recommended_plugins',
    '--disable', 'in_app_browser', '--enable', 'skip_host_skill_discovery',
    '-c', 'cli_auth_credentials_store="file"', '-c', 'forced_login_method="chatgpt"',
    'app-server', '--strict-config', '--stdio'];
}
function prepare(files) {
  // Fixed discovery roots must be absent in the reviewed image. The engine
  // home is exclusively created, never adopted, reset or reused on a new turn.
  for (const candidate of ['/etc/codex', '/.codex', '/AGENTS.md', '/tmp/.codex', '/tmp/AGENTS.md']) {
    if (files.existsSync(candidate)) throw new Error('Engine configuration refused');
  }
  // Read exactly one bounded private frame. Never buffer ahead: the remaining
  // stdin belongs to the app-server protocol after this function returns.
  const magic = Buffer.from('COMMONROOM-AUTH-V1\n');
  function exact(size) {
    const data = Buffer.alloc(size);
    let offset = 0;
    while (offset < size) {
      const n = files.readSync(0, data, offset, size - offset, null);
      if (!Number.isInteger(n) || n <= 0 || n > size - offset) throw new Error('Engine authentication refused');
      offset += n;
    }
    return data;
  }
  if (!exact(magic.length).equals(magic)) throw new Error('Engine authentication refused');
  const size = exact(4).readUInt32BE(0);
  if (size < 2 || size > 65536) throw new Error('Engine authentication refused');
  const raw = exact(size);
  const auth = JSON.parse(raw.toString('utf8'));
  if (!auth || auth.auth_mode !== 'chatgpt' ||
      (auth.OPENAI_API_KEY !== undefined && auth.OPENAI_API_KEY !== null) ||
      !auth.tokens || typeof auth.tokens.access_token !== 'string' || !auth.tokens.access_token)
    throw new Error('Subscription authentication required');
  files.mkdirSync(HOME, {mode: 0o700});
  files.mkdirSync(CODEX_HOME, {mode: 0o700});
  // Publish only validated bytes, with exclusive creation and private mode.
  files.writeFileSync(`${CODEX_HOME}/auth.json`, raw, {flag: 'wx', mode: 0o600});
  files.writeFileSync(`${CODEX_HOME}/config.toml`,
    'cli_auth_credentials_store = "file"\nforced_login_method = "chatgpt"\n',
    {flag: 'wx', mode: 0o600});
}

// Read only the final private inode, after the reviewed provider child closed.
// No path, credential or executable is supplied by a model or socket request.
function finalCredential(files) {
  const fd = files.openSync(`${CODEX_HOME}/auth.json`, files.constants.O_RDONLY | files.constants.O_NOFOLLOW | files.constants.O_NONBLOCK);
  try {
    const st = files.fstatSync(fd);
    if (!st.isFile() || st.uid !== 1000 || st.nlink !== 1 || (st.mode & 0o077) !== 0 || st.size < 2 || st.size > 65536)
      throw Error('Credential export refused');
    const raw = Buffer.alloc(65537);
    let size = 0;
    while (size < raw.length) {
      const n = files.readSync(fd, raw, size, raw.length - size, null);
      if (!n) break;
      size += n;
    }
    if (size < 2 || size > 65536) throw Error('Credential export refused');
    const auth = JSON.parse(raw.subarray(0, size).toString('utf8'));
    if (auth?.auth_mode !== 'chatgpt' || (auth.OPENAI_API_KEY !== undefined && auth.OPENAI_API_KEY !== null) ||
        typeof auth.tokens?.access_token !== 'string' || !auth.tokens.access_token) throw Error('Credential export refused');
    return raw.subarray(0, size);
  } finally { files.closeSync(fd); }
}

// The worker alone enters this fixed, one-shot private channel using docker exec
// on the inspected engine ID. It is never a model tool or app-server response.
// The child must close before export; this is NOT container/descendant absence.
function credentialExporter({files, network, child, fail}) {
  const magic = Buffer.from('COMMONROOM-CREDENTIAL-EXPORT-V1\n');
  let requested = false, completed = false, closed = false, client;
  const server = network.createServer({allowHalfOpen: true}, socket => {
    if (closed || requested || client) { socket.destroy(); return; }
    client = socket;
    let input = Buffer.alloc(0);
    socket.setTimeout(10000, () => { socket.destroy(); fail(); });
    socket.on('error', () => fail());
    socket.on('data', chunk => {
      if (input.length + chunk.length > magic.length) { socket.destroy(); fail(); return; }
      input = Buffer.concat([input, chunk]);
    });
    socket.once('end', () => {
      if (closed || requested || !input.equals(magic)) { socket.destroy(); fail(); return; }
      requested = true;
      try { if (!child.kill('SIGTERM')) { socket.destroy(); fail(); } }
      catch (_) { socket.destroy(); fail(); }
    });
  });
  child.once('close', () => {
    if (!requested || closed) return;
    try {
      const raw = finalCredential(files), size = Buffer.alloc(4);
      size.writeUInt32BE(raw.length);
      client.end(Buffer.concat([magic, size, raw]));
      raw.fill(0);
      completed = true;
    } catch (_) { fail(); }
  });
  server.on('error', fail);
  server.listen(`${HOME}/credential-export.sock`);
  return {requested: () => requested, complete: () => completed,
    close() { closed = true; try { client?.destroy(); } catch (_) {} try { server.close(); } catch (_) {} }};
}

// Injection is synthetic-test infrastructure, never a model/CLI configuration.
// Signal forwarding is not descendant/container shutdown proof; the supervisor
// must still remove and observe both immutable container identities on exit.
function run({files = fs, network = net, processes = cp, proc = process,
  stderr = process.stderr, exportFactory = credentialExporter} = {}) {
  const report = () => { try { stderr.write('Reviewed engine startup refused\n'); } catch (_) {} };
  if (proc.platform !== 'linux' || typeof proc.getuid !== 'function' || proc.getuid() !== 1000 || !Array.isArray(proc.argv) || proc.argv.length !== 2) {
    report(); return Promise.resolve(70);
  }
  try { prepare(files); } catch (_) { report(); return Promise.resolve(70); }
  return new Promise(resolve => {
    let server, child, exporter, settled = false;
    const sockets = new Set();
    const remember = socket => { sockets.add(socket); socket.on('close', () => sockets.delete(socket)); return socket; };
    const interrupt = () => { try { child?.kill('SIGINT'); } catch (_) {} };
    const terminate = () => { if (exporter?.complete()) { finish(0); return; } try { child?.kill('SIGTERM'); } catch (_) {} };
    const finish = (code, failed = false) => {
      if (settled) return;
      settled = true;
      proc.removeListener('SIGINT', interrupt); proc.removeListener('SIGTERM', terminate);
      if (failed) { terminate(); report(); }
      for (const socket of sockets) { try { socket.destroy(); } catch (_) {} }
      try { server?.close(); } catch (_) {}
      try { exporter?.close(); } catch (_) {}
      resolve(code);
    };
    try {
      server = network.createServer(client => {
        if (settled) { client.destroy(); return; }
        remember(client);
        let remote;
        try { remote = remember(network.connect('/run/provider.sock')); }
        catch (_) { client.destroy(); return; }
        client.pipe(remote); remote.pipe(client);
        client.on('error', () => remote.destroy()); remote.on('error', () => client.destroy());
        client.on('close', () => remote.destroy()); remote.on('close', () => client.destroy());
      });
      server.on('error', () => finish(70, true));
      server.listen(18080, '127.0.0.1', () => {
        if (settled) return;
        try {
          child = processes.spawn('/opt/commonroom/bin/codex', invocation(), {
            shell: false, cwd: HOME, env: environment(), stdio: ['inherit', 'inherit', 'ignore']
          });
          proc.on('SIGINT', interrupt); proc.on('SIGTERM', terminate);
          child.once('error', () => finish(70, true));
          child.once('close', (code, signal) => {
            if (settled) return;
            if (exporter?.requested()) {
              for (const socket of sockets) { try { socket.destroy(); } catch (_) {} }
              try { server.close(); } catch (_) {}
              return; // Remain alive for private export and supervisor fencing.
            }
            if (signal || code !== 0) report();
            finish(!signal && code === 0 ? 0 : 70);
          });
          exporter = exportFactory({files, network, child, fail: () => finish(70, true)});
        } catch (_) { finish(70, true); }
      });
    } catch (_) { finish(70, true); }
  });
}
module.exports = {environment, invocation, prepare, finalCredential, credentialExporter, run};
if (require.main === module) run().then(code => { process.exitCode = code; }, () => { process.exitCode = 70; });
