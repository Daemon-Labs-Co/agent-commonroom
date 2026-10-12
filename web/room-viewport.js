// A bounded desktop room layout keeps the composer in view. Mobile retains
// natural document scrolling; only message/squad panels scroll on desktop.
function fitRoomViewport(){
 const enabled=state?.featureFlags?.roomViewportFit!==false&&view==='Room'&&!!document.querySelector('#send');
 const entering=enabled&&!document.body.classList.contains('room-viewport-fit');
 document.body.classList.toggle('room-viewport-fit',enabled);
 if(entering)window.scrollTo({top:0,left:0,behavior:'instant'});
}
const roomViewportRender=render;render=function(){roomViewportRender();fitRoomViewport()};

fitRoomViewport();
