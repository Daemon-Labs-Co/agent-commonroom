// Page alignment only; preserve view state, actions and scroll containers.
function alignApplicationViews(){document.body.classList.toggle('top-aligned-views',state?.featureFlags?.topAlignedViews!==false)}
const alignedViewsRender=render;render=function(){alignedViewsRender();alignApplicationViews()};
alignApplicationViews();
