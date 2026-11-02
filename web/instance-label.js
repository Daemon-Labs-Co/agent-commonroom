'use strict';
// Static-only previews cannot determine the application machine's hostname.
globalThis.commonroomServerHostname = 'Server name unavailable';
document.querySelector('#instance-address').textContent = globalThis.commonroomServerHostname;
