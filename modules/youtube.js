(function() {

  module.exports = function(robot) {
    return robot.respond(/(youtube|yt)( me)? (.*)/i, function(msg) {
      var query;
      query = msg.match[3];
      return msg.http("http://gdata.youtube.com/feeds/api/videos").query({
        orderBy: "relevance",
        'max-results': 15,
        alt: 'json',
        q: query
      }).get()(function(err, res, body) {
        var video, videos;
        videos = JSON.parse(body);
        videos = videos.feed.entry;
        video = msg.random(videos);
        return video.link.forEach(function(link) {
          if (link.rel === "alternate" && link.type === "text/html") {
            return msg.send(link.href);
          }
        });
      });
    });
  };

}).call(this);
