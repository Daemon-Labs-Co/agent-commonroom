'use strict';
// The container has no external network. This loopback bridge exposes only the
// host's per-turn, provider-allowlisted Unix gateway; it is not a Docker socket.
const net=require('node:net'),fs=require('node:fs'),path=require('node:path'),cp=require('node:child_process');
const [provider,...args]=process.argv.slice(2);
if(!['codex','claude'].includes(provider))process.exit(64);
const dir=path.join('/home/agent',provider==='codex'?'.codex':'.claude');fs.mkdirSync(dir,{recursive:true,mode:0o700});
if(args[0]!=='auth-login')fs.copyFileSync('/run/provider-auth.json',path.join(dir,provider==='codex'?'auth.json':'.credentials.json'));if(args[0]!=='auth-login')fs.chmodSync(path.join(dir,provider==='codex'?'auth.json':'.credentials.json'),0o600);
if(args[0]==='auth-login'){args.splice(0,args.length,...(provider==='codex'?['login','--device-auth']:['auth','login','--claudeai']));}
const server=net.createServer(client=>{const remote=net.connect('/run/provider.sock');client.pipe(remote);remote.pipe(client);client.on('error',()=>remote.destroy());remote.on('error',()=>client.destroy());client.on('close',()=>remote.destroy())});
server.listen(18080,'127.0.0.1',()=>{const child=cp.spawn(provider,args,{stdio:'inherit',shell:false});child.on('error',()=>{server.close();process.exit(70)});child.on('exit',(code)=>{server.close();process.exit(code??70)});for(const signal of ['SIGTERM','SIGINT'])process.on(signal,()=>child.kill(signal))});
server.on('error',()=>process.exit(70));
