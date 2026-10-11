"use strict";
let catalog=[];
const escapeHTML=value=>String(value).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
function destination(value,entry,image=false){
 try{
  const url=new URL(value,'https://documentation.invalid/'+entry.path);
  if(url.origin==='https://documentation.invalid'){
   const path=decodeURIComponent(url.pathname.slice(1));
   if(image){const name=path.split('/').pop();return ['docs/screenshots/room.jpg','docs/screenshots/workflow.jpg','docs/screenshots/connections.jpg'].includes(path)?'/help/image/'+name:null}
   const found=catalog.find(item=>item.path===path);return found?'/help/?doc='+encodeURIComponent(found.id)+url.hash:null;
  }
  return !image&&['http:','https:'].includes(url.protocol)?url.href:null;
 }catch{return null}
}
function inline(value,entry){
 // Parse links/code before escaping. Every output fragment escapes user text;
 // raw HTML and executable URL schemes are deliberately unsupported.
 const pattern=/(!?\[([^\]]*)\]\(([^)]+)\)|`([^`]+)`|\*\*([^*]+)\*\*)/g;
 let html='',last=0;
 for(const match of value.matchAll(pattern)){
  html+=escapeHTML(value.slice(last,match.index));last=match.index+match[0].length;
  if(match[4]!==undefined)html+='<code>'+escapeHTML(match[4])+'</code>';
  else if(match[5]!==undefined)html+='<strong>'+escapeHTML(match[5])+'</strong>';
  else{
   const image=match[0].startsWith('!'),url=destination(match[3],entry,image);
   if(!url)html+=escapeHTML(match[2]);
   else if(image)html+='<img src="'+escapeHTML(url)+'" alt="'+escapeHTML(match[2])+'" loading="lazy">';
   else html+='<a href="'+escapeHTML(url)+'"'+(url.startsWith('/help/')?'':' target="_blank" rel="noopener noreferrer"')+'>'+escapeHTML(match[2])+'</a>';
  }
 }
 return html+escapeHTML(value.slice(last));
}
function markdown(text,entry){
 const lines=text.replace(/\r\n/g,'\n').split('\n');let html='',paragraph=[],list=null;
 const flush=()=>{if(paragraph.length){html+='<p>'+inline(paragraph.join(' '),entry)+'</p>';paragraph=[]}if(list){html+='</'+list+'>';list=null}};
 const slugs=new Map();
 for(let i=0;i<lines.length;i++){
  const line=lines[i];
  if(line.startsWith('```')){flush();let code=[];while(++i<lines.length&&!lines[i].startsWith('```'))code.push(lines[i]);html+='<pre><code>'+escapeHTML(code.join('\n'))+'</code></pre>';continue}
  const heading=line.match(/^(#{1,6})\s+(.+)$/);
  if(heading){flush();const base=heading[2].toLowerCase().replace(/[^\p{L}\p{N}\s-]/gu,'').replace(/\s/g,'-');const count=slugs.get(base)||0;slugs.set(base,count+1);const slug=base+(count?'-'+count:'');html+='<h'+heading[1].length+' id="'+escapeHTML(slug)+'">'+inline(heading[2],entry)+'</h'+heading[1].length+'>';continue}
  if(line.startsWith('|')&&i+1<lines.length&&/^\|[\s:|\-]+\|$/.test(lines[i+1])){
   flush();const cells=row=>row.trim().replace(/^\||\|$/g,'').split('|').map(x=>x.trim());html+='<table><thead><tr>'+cells(line).map(x=>'<th>'+inline(x,entry)+'</th>').join('')+'</tr></thead><tbody>';i++;
   while(i+1<lines.length&&lines[i+1].startsWith('|'))html+='<tr>'+cells(lines[++i]).map(x=>'<td>'+inline(x,entry)+'</td>').join('')+'</tr>';
   html+='</tbody></table>';continue;
  }
  const item=line.match(/^\s*(?:([-*])|\d+\.)\s+(.+)$/);
  if(item){if(paragraph.length)flush();const kind=item[1]?'ul':'ol';if(list!==kind){flush();html+='<'+kind+'>';list=kind}html+='<li>'+inline(item[2],entry)+'</li>';continue}
  if(!line.trim()){flush();continue}if(list)flush();paragraph.push(line);
 }
 flush();return html;
}
function showGuides(){
 const search=document.querySelector('#find-guide').value.toLowerCase();const current=new URLSearchParams(location.search).get('doc')||'index';
 document.querySelector('#guides').innerHTML=catalog.filter(e=>e.title.toLowerCase().includes(search)).map(e=>'<a href="/help/?doc='+encodeURIComponent(e.id)+'"'+(e.id===current?' aria-current="page"':'')+'>'+escapeHTML(e.title)+'</a>').join('')||'<p>No matching guides.</p>';
}
(async()=>{
 const status=document.querySelector('#help-status');
 try{
  const manifest=await fetch('/help/catalog');if(!manifest.ok)throw Error('Documentation catalog unavailable.');catalog=await manifest.json();showGuides();document.querySelector('#find-guide').addEventListener('input',showGuides);
  const id=new URLSearchParams(location.search).get('doc')||'index';const entry=catalog.find(e=>e.id===id);if(!entry)throw Error('Guide not found. Choose a guide from the menu.');
  const response=await fetch('/help/document/'+encodeURIComponent(id));if(!response.ok)throw Error('This guide is missing from the installation. Deploy docs and README files alongside the application.');
  const text=await response.text();document.querySelector('#document').innerHTML=id==='license'?'<pre>'+escapeHTML(text)+'</pre>':markdown(text,entry);status.textContent='';document.title=entry.title+' · Agent CommonRoom Help';
  if(location.hash){try{document.getElementById(decodeURIComponent(location.hash.slice(1)))?.scrollIntoView()}catch{}}
 }catch(error){status.textContent=error.message}
})();
