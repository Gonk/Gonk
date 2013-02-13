(function() {
  module.exports = function(robot) {
    robot.respond(/(image|img)( me)? (.*)/i, function(msg) {
      return imageMe(msg, msg.match[3], function(url) {
        return msg.send(url);
      });
    });
    robot.respond(/animate( me)? (.*)/i, function(msg) {
      return imageMe(msg, msg.match[2], true, function(url) {
        return msg.send(url);
      });
    });
    return robot.respond(/(?:mo?u)?sta(?:s|c)he?(?: me)? (.*)/i, function(msg) {
      var imagery, mustachify, type;
      type = Math.floor(Math.random() * 3);
      mustachify = "http://mustachify.me/" + type + "?src=";
      imagery = msg.match[1];
      if (imagery.match(/^https?:\/\//i)) {
        return msg.send("" + mustachify + imagery);
      } else {
        return imageMe(msg, imagery, false, true, function(url) {
          return msg.send("" + mustachify + url);
        });
      }
    });
  };

  imageMe = function(msg, query, animated, faces, cb) {
    var q;
    if (typeof animated === 'function') cb = animated;
    if (typeof faces === 'function') cb = faces;
    q = {
      v: '1.0',
      rsz: '8',
      q: query,
      safe: 'active'
    };
    if (typeof animated === 'boolean' && animated === true) q.as_filetype = 'gif';
    if (typeof faces === 'boolean' && faces === true) q.imgtype = 'face';
    return msg.http('http://ajax.googleapis.com/ajax/services/search/images').query(q).get()(function(err, res, body) {
      var image, images, _ref;
      images = JSON.parse(body);
      images = (_ref = images.responseData) != null ? _ref.results : void 0;
      if ((images != null ? images.length : void 0) > 0) {
        image = msg.random(images);
        return cb("" + image.unescapedUrl + "#.png");
      }
    });
  };

}).call(this);
