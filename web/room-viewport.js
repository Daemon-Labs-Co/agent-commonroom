// A bounded desktop room layout keeps the composer in view. Mobile retains
// natural document scrolling; only message/squad panels scroll on desktop.
function fitRoomViewport(){
 const enabled=state?.featureFlags?.roomViewportFit!==false&&view==='Room'&&!!document.querySelector('#send');
 const entering=enabled&&!document.body.classList.contains('room-viewport-fit');
 document.body.classList.toggle('room-viewport-fit',enabled);
 if(entering)window.scrollTo({top:0,left:0,behavior:'instant'});
}
// Base rendering temporarily enlarges the panel before tabs/task cards are
// inserted. Restoring there clamps scrollTop too early. Capture a visible message
// and restore only after every inner decorator and the final room layout.
let roomViewportScrollKey=null;
function captureRoomScroll(list){
 if(!list||typeof list.scrollTop!=='number')return null;
 const top=list.getBoundingClientRect?.().top;
 const rows=[...(list.querySelectorAll?.(':scope > .message')||[])];
 const anchor=rows.find(row=>row.dataset?.messageId&&row.getBoundingClientRect().bottom>top);
 return {top:list.scrollTop,atBottom:list.scrollHeight-list.clientHeight-list.scrollTop<=24,atTop:list.scrollTop<=24,anchor:anchor?{id:anchor.dataset.messageId,offset:anchor.getBoundingClientRect().top-top}:null};
}
function restoreRoomScroll(list,snapshot,fresh,newestFirst){
 if(!list||typeof list.scrollTop!=='number')return;
 if(fresh||!snapshot||(newestFirst?snapshot.atTop:snapshot.atBottom)){list.scrollTop=newestFirst?0:list.scrollHeight;return}
 const anchor=snapshot.anchor&&[...(list.querySelectorAll?.(':scope > .message')||[])].find(row=>row.dataset?.messageId===snapshot.anchor.id);
 if(anchor){list.scrollTop+=anchor.getBoundingClientRect().top-list.getBoundingClientRect().top-snapshot.anchor.offset}
 else list.scrollTop=snapshot.top;
}
const roomViewportRender=render;render=function(){
 const enabled=state?.featureFlags?.roomViewportFit!==false&&view==='Room';
 const snapshot=enabled?captureRoomScroll(document.querySelector('.message-list')):null;
 const newestFirst=state?.featureFlags?.newestMessagesFirst===true;
 const key=(typeof roomId==='undefined'?'':roomId)+':'+newestFirst+':'+(state?.messages?.at(-1)?.id||'');
 roomViewportRender();fitRoomViewport();
 if(enabled){restoreRoomScroll(document.querySelector('.message-list'),snapshot,key!==roomViewportScrollKey,newestFirst);roomViewportScrollKey=key}
 else roomViewportScrollKey=null;
};

fitRoomViewport();
