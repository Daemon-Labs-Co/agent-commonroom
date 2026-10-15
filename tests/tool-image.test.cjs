'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const {spawnSync} = require('node:child_process');
const root = path.resolve(__dirname, '..');
const recipe = fs.readFileSync(path.join(root, 'docker/tools/Dockerfile'), 'utf8');
const ignore = fs.readFileSync(path.join(root, 'docker/tools/Dockerfile.dockerignore'), 'utf8');

test('tool build context deny-all with only reviewed source exceptions', () => {
  const rules = ignore.split(/\r?\n/).map(s => s.trim()).filter(s => s && !s.startsWith('#'));
  assert.equal(rules[0], '**');
  const subtreeExclusions = new Set(['cmd/**','cmd/tool-helper/**','cmd/workspace-bootstrap/**','internal/**','internal/toolworkspace/**','internal/workspacebootstrap/**','internal/strictjson/**','docker/**','docker/tools/**']);
  const allowed = new Set([
    'go.mod','cmd/','cmd/tool-helper/','cmd/tool-helper/main.go','cmd/workspace-bootstrap/','cmd/workspace-bootstrap/main.go',
    'internal/','internal/toolworkspace/','internal/toolworkspace/workspace_unix.go',
    'internal/strictjson/','internal/strictjson/encoding.go','internal/workspacebootstrap/','internal/workspacebootstrap/bootstrap.go',
    'docker/','docker/tools/','docker/tools/Dockerfile','docker/tools/tool-call.cjs'
  ]);
  for (const rule of rules.slice(1)) {
    if (subtreeExclusions.has(rule)) { subtreeExclusions.delete(rule); continue; }
    assert.ok(rule.startsWith('!') && allowed.has(rule.slice(1)), 'unreviewed context exception');
    allowed.delete(rule.slice(1));
  }
  assert.equal(allowed.size, 0);
  assert.equal(subtreeExclusions.size, 0);
  // Parent exceptions must be followed by subtree exclusion before leaf opt-in.
  for (const dir of ['cmd','cmd/tool-helper','cmd/workspace-bootstrap','internal','internal/toolworkspace','internal/strictjson','internal/workspacebootstrap','docker','docker/tools']) {
    assert.ok(rules.indexOf('!'+dir+'/') < rules.indexOf(dir+'/**'));
    const leafExceptions = rules.filter(rule => rule.startsWith('!'+dir+'/') && !rule.endsWith('/'));
    assert.ok(leafExceptions.every(rule => rules.indexOf(rule) > rules.indexOf(dir+'/**')));
  }
});

test('image copies only declared tool source and helper artifact, never whole tree', () => {
  const copies = recipe.split(/\r?\n/).filter(line => /^COPY\s/.test(line));
  assert.deepEqual(copies, [
    'COPY go.mod ./go.mod',
    'COPY cmd/tool-helper/main.go ./cmd/tool-helper/main.go',
    'COPY cmd/workspace-bootstrap/main.go ./cmd/workspace-bootstrap/main.go',
    'COPY internal/workspacebootstrap/bootstrap.go ./internal/workspacebootstrap/bootstrap.go',
    'COPY internal/toolworkspace/workspace_unix.go ./internal/toolworkspace/workspace_unix.go',
    'COPY internal/strictjson/encoding.go ./internal/strictjson/encoding.go',
    'COPY --from=build --chmod=0555 /out/tool-helper /opt/commonroom/tool-helper',
    'COPY --from=build --chmod=0555 /out/workspace-bootstrap /opt/commonroom/workspace-bootstrap',
    'COPY --chmod=0444 docker/tools/tool-call.cjs /opt/commonroom/tool-call.cjs'
  ]);
  const instructions = recipe.split(/\r?\n/).filter(s => s && !s.startsWith('#')).join('\n');
  assert.doesNotMatch(instructions, /\bADD\s|npm\s|apt-get\s|--mount=|--secret=|--ssh=/);
  assert.match(instructions, /RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off/);
  assert.match(instructions, /USER 1000:1000/);
  assert.match(instructions, /test -x \/usr\/bin\/git/);
  assert.match(instructions, /ENTRYPOINT \["\/usr\/local\/bin\/node", "\/opt\/commonroom\/tool-call.cjs"\]/);
});

test('both base images require external digest pins, no defaults', () => {
  assert.match(recipe, /^ARG GO_IMAGE$/m);
  assert.match(recipe, /^ARG NODE_IMAGE$/m);
  assert.doesNotMatch(recipe, /^ARG (?:GO_IMAGE|NODE_IMAGE)=/m);
  assert.match(recipe, /for image in "\$GO_IMAGE" "\$NODE_IMAGE"/);
  assert.match(recipe, /test "\$\{#digest\}" -eq 64/);
  assert.match(recipe, /\*\[!0-9a-f\]\*/);
});

test('actual recipe pin guard refuses tags, missing, malformed and uppercase digests', () => {
  const match = recipe.match(/RUN (for image[\s\S]*?)\nWORKDIR/);
  assert.ok(match, 'pin validation block missing');
  const script = match[1].replace(/\\\n/g, '\n');
  const pinned = 'synthetic/base@sha256:' + 'a'.repeat(64);
  for (const [goImage,nodeImage,success] of [
    [pinned,pinned,true],['synthetic:latest',pinned,false],
    [pinned,'synthetic:latest',false],['',pinned,false],
    [pinned,pinned.slice(0,-1),false],
    [pinned,'synthetic/base@sha256:'+'A'.repeat(64),false]
  ]) {
    // Only the trusted recipe's shell validation is executed. No Docker/build,
    // providers, downloads or filesystem fixture data are involved.
    const result = spawnSync('/bin/sh', ['-c', script], {
      env: {GO_IMAGE:goImage,NODE_IMAGE:nodeImage,PATH:'/usr/bin:/bin'},
      timeout:1000, stdio:'pipe'
    });
    assert.equal(result.status === 0, success);
    assert.equal(result.error, undefined);
  }
});
