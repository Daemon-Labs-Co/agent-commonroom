// Optional display-only order. Append existing rendered nodes after all avatar
// decorators run; never sort/mutate messages, audit events, timestamps or prompts.
function arrangeNewestRoomMessages(list){
 const rows=[...list.querySelectorAll(':scope > .message')];
 for(const row of rows.reverse())list.appendChild(row);
}
let roomMessageOrderKey=null;
const roomMessageOrderRender=render;
render=function(){
 const oldScroll=document.querySelector('.message-list')?.scrollTop;
 roomMessageOrderRender();
 const list=document.querySelector('.message-list');
 if(view!=='Room'||!list||state?.featureFlags?.newestMessagesFirst!==true){roomMessageOrderKey=null;return}
 const key=roomId+':'+(state.messages.at(-1)?.id||'');
 arrangeNewestRoomMessages(list);
 list.scrollTop=key!==roomMessageOrderKey||oldScroll===undefined?0:oldScroll;
 roomMessageOrderKey=key;
};
