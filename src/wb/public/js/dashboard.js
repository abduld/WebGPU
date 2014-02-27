$(function() {


  if (!Modernizr.websockets) {
    bootbox.alert({
      'title': 'Unsupported Browser',
      'message':
        '<p>' +
          'This page requires a recent web browser with websocket support.' +
        '</p>'
    });
    return;
  }

  var stripHTTP = function(str) {
    return str.replace('http://', '');
  };

  var addr = 'ws://' + stripHTTP(domainname) + '/websocket/dashboard';

  console.log(addr);

  window.socket = new WebSocket(addr);

  var safeJSONParse = function(s) {
    var res;
    try {
      res = JSON.parse(s);
    } catch (e) {
      res = s;
    }
    return res;
  }

  socket.onmessage = function(evt) {
    var msg = safeJSONParse(evt.data);
    var data = safeJSONParse(safeJSONParse(msg.data));
    if (_.has(data, 'message')) {
      var obj = safeJSONParse(data.message);
      if (_.isArray(obj) && obj.length == 1) {
        obj = obj[0];
      }
      data.message = obj;
    }
    msg.data = data;
    render(msg);
  };

  var render = function(obj) {
    console.log(obj);
  };
});

