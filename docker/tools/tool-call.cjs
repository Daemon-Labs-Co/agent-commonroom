'use strict';

// Fixed entrypoint for the planned credential-free tool image. This is not a
// host tool API. Source changes CAN expose host tools/files, BUT invalidate the
// mandatory isolation contract. Broker admission and verified mounts remain
// required; argv, input and inherited env never select executables or roots.
const {spawn} = require('node:child_process');

function helperEnvironment() {
  return {
    HOME: '/tmp/tool-home',
    PATH: '/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin',
    HTTP_PROXY: '', HTTPS_PROXY: '', ALL_PROXY: '', NO_PROXY: '',
    NODE_OPTIONS: '', NODE_USE_ENV_PROXY: '0', NODE_VERSION: '', YARN_VERSION: ''
  };
}

// Dependencies are injected only by synthetic tests, never by CLI/model input.
// The outer supervisor must prove container shutdown; signal forwarding alone
// is not proof that descendants stopped or an ambiguous write did not occur.
function run({cp = {spawn}, proc = process, stderr = process.stderr} = {}) {
  const report = () => { try { stderr.write('Tool helper launch refused\n'); } catch (_) {} };
  if (proc.platform !== 'linux' || !Array.isArray(proc.argv) || proc.argv.length !== 2) {
    report();
    return Promise.resolve(2);
  }
  return new Promise(resolve => {
    let child;
    try {
      child = cp.spawn('/opt/commonroom/tool-helper', [], {
        shell: false, cwd: '/workspace/repository', env: helperEnvironment(),
        stdio: 'inherit'
      });
    } catch (_) {
      report(); resolve(2); return;
    }
    let settled = false;
    const interrupt = () => { try { child.kill('SIGINT'); } catch (_) {} };
    const terminate = () => { try { child.kill('SIGTERM'); } catch (_) {} };
    const finish = code => {
      if (settled) return;
      settled = true;
      proc.removeListener('SIGINT', interrupt);
      proc.removeListener('SIGTERM', terminate);
      resolve(code);
    };
    proc.on('SIGINT', interrupt);
    proc.on('SIGTERM', terminate);
    child.once('error', () => { report(); finish(2); });
    child.once('close', (code, signal) => {
      finish(!signal && [0, 2, 3].includes(code) ? code : 2);
    });
  });
}

module.exports = {run};
if (require.main === module) {
  run().then(code => { process.exitCode = code; }, () => { process.exitCode = 2; });
}
