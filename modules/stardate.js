(function() {

  module.exports = function(robot) {
    return robot.respond(/stardate.*(January|February|March|April|May|June|July|August|September|October|November|December).*(\d+).*,? (\d+)/i, function(msg) {
      return msg.http('http://www.stoacademy.com/tools/show_stardate.php')
      .header('Content-Type', 'application/x-www-form-urlencoded')
      .query({
        tostardate: true,
        month: msg.match[1],
        date: msg.match[2],
        year: msg.match[3],
        hour: 0,
        minutes: 0
      })
      .post()(function(err, res, body) {
        splits = body.split(" ");
        return msg.send("Stardate: " + splits[splits.length - 1]);
      });
    });
  };

}).call(this);
