const assert=require('node:assert/strict'),fs=require('node:fs'),vm=require('node:vm');
const context=vm.createContext({render(){},Date,esc:s=>String(s).replaceAll('&','&amp;').replaceAll('<','&lt;').replaceAll('>','&gt;')});
vm.runInContext(fs.readFileSync('web/connection-testing.js','utf8'),context);
const result=vm.runInContext(`connectionTestMarkup({status:'needs-sign-in',checkedAt:'2026-10-11T00:00:00Z',checks:[{name:'<script>',status:'required',detail:'<img src=x>'}]})`,context);
assert.ok(result.includes('Sign-in needed'));assert.ok(result.includes('&lt;script&gt;'));assert.ok(!result.includes('<img'));
const ready=vm.runInContext(`connectionTestMarkup({status:'working',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(ready.includes('Connection working'));
console.log('Connection readiness labels and escaped diagnostic rendering passed.');

const failed=vm.runInContext(`connectionTestMarkup({status:'failed',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(failed.includes('Connection test failed'));
const unverified=vm.runInContext(`connectionTestMarkup({status:'ready',checkedAt:'2026-10-11T00:00:00Z',checks:[]})`,context);assert.ok(!unverified.includes('Connection working'));assert.ok(unverified.includes('provider unverified'));
