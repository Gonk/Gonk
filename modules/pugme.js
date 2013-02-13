(function() {

  module.exports = function(robot) {
    robot.respond(/pug me/i, function(msg) {
      return msg.http("http://pugme.herokuapp.com/random").get()(function(err, res, body) {
        return msg.send(JSON.parse(body).pug);
      });
    });
    robot.respond(/pug bomb( (\d+))?/i, function(msg) {
      var count;
      count = msg.match[2] || 5;
      count = Math.min(count, 15); // Keep the number of pugs sane
      return msg.http("http://pugme.herokuapp.com/bomb?count=" + count)
      .headers({
        'Accept-Language': 'en-us,en;q=0.5',
        'Accept-Charset': 'utf-8',
        'User-Agent': "Mozilla/5.0 (X11; Linux x86_64; rv:2.0.1) Gecko/20100101 Firefox/4.0.1"
      }).get()(function(err, res, body) {
        var pug, _i, _len, _ref, _results;
        _ref = JSON.parse(body).pugs;
        return msg.send(_ref.join(' '));
      });
    });
    return robot.respond(/how many pugs are there/i, function(msg) {
      return msg.http("http://pugme.herokuapp.com/count").get()(function(err, res, body) {
        return msg.send("There are " + (JSON.parse(body).pug_count) + " pugs.");
      });
    });
  };

}).call(this);
