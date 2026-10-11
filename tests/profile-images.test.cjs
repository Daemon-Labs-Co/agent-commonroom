const {test}=require('node:test');const assert=require('node:assert/strict');const fs=require('node:fs');const vm=require('node:vm');
const source=fs.readFileSync('web/profile-images.js','utf8');const scope={form(){},render(){}};vm.createContext(scope);vm.runInContext(source,scope);
test('photo presentation rejects external or active content and respects hidden flag',()=>{
 const valid={image:'data:image/jpeg;base64,YQ=='};assert.equal(scope.humanPhoto(valid),valid.image);assert.equal(scope.humanPhoto(valid,false),'');
 for(const image of ['https://example.com/p.jpg','data:image/svg+xml;base64,YQ==','javascript:alert(1)','data:image/jpeg;base64,\" onerror=evil'])assert.equal(scope.humanPhoto({image}),'');
 assert.equal(scope.humanPhoto(null),'');
});
test('upload rejects unsupported formats and excessive size before decoding',async()=>{
 await assert.rejects(scope.resizeHumanPhoto({type:'image/svg+xml',size:10}),/JPEG or PNG/);
 await assert.rejects(scope.resizeHumanPhoto({type:'image/png',size:6*1024*1024}),/5 MB/);
});
