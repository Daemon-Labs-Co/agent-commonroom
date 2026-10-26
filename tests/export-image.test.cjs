'use strict';
const {test} = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(__dirname, '..');
const recipe = fs.readFileSync(path.join(root,'docker/export/Dockerfile'),'utf8');
const rules = fs.readFileSync(path.join(root,'docker/export/Dockerfile.dockerignore'),'utf8').trim().split('\n');

test('probe context admits only reviewed credential-free helper sources',()=>{
 assert.deepEqual(rules, ['**','!go.mod','!cmd/','cmd/**','!cmd/workspace-mount-probe/','cmd/workspace-mount-probe/**','!cmd/workspace-mount-probe/main.go','!internal/','internal/**','!internal/workspacemount/','internal/workspacemount/**','!internal/workspacemount/proof.go','!internal/workspacemount/observe_linux.go','!docker/','docker/**','!docker/export/','docker/export/**','!docker/export/Dockerfile']);
 assert.deepEqual(recipe.split('\n').filter(s=>s.startsWith('COPY ')),[
 'COPY go.mod ./go.mod',
 'COPY cmd/workspace-mount-probe/main.go ./cmd/workspace-mount-probe/main.go',
 'COPY internal/workspacemount/proof.go ./internal/workspacemount/proof.go',
 'COPY internal/workspacemount/observe_linux.go ./internal/workspacemount/observe_linux.go',
 'COPY --from=build /rootfs/ /',
 'COPY --from=build --chmod=0555 /out/workspace-mount-probe /opt/commonroom/workspace-mount-probe']);
});
test('probe runtime is scratch with one fixed executable and no credentials or downloads',()=>{
 assert.match(recipe,/^ARG GO_IMAGE$/m);
 assert.doesNotMatch(recipe,/^ARG GO_IMAGE=/m);
 assert.match(recipe,/test "\$\{#digest\}" -eq 64/);
 assert.match(recipe,/\*\[!0-9a-f\]\*/);
 assert.match(recipe,/RUN --network=none CGO_ENABLED=0 GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off/);
 const runtime=recipe.split('FROM scratch\n')[1];assert.ok(runtime);
 assert.deepEqual(runtime.trim().split('\n'),[
 'COPY --from=build /rootfs/ /',
 'COPY --from=build --chmod=0555 /out/workspace-mount-probe /opt/commonroom/workspace-mount-probe',
 'USER 1000:1000','WORKDIR /workspace','ENTRYPOINT ["/opt/commonroom/workspace-mount-probe"]']);
 assert.doesNotMatch(recipe,/\bADD\s|apt-get|npm\s|--secret=|--ssh=|--mount=/);
});
