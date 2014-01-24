(function() {
  module.exports = function(robot) {
    return robot.hear(/\/r\/(.*)/i, function(msg) {
      return msg.send('http://reddit.com/r/' + msg.match[1]);
    });
  }
}).call(this);
