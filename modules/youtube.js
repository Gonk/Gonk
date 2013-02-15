(function() {

  module.exports = function(robot) {
    return robot.respond(/(youtube|yt)( me)? (.*)/i, function(msg) {
      var query;
      query = msg.match[3];
      return msg
      .http("http://gdata.youtube.com/feeds/api/videos")
      .query({
        orderBy: "relevance",
        'max-results': 1, // Get the best match so queries are consistent
        alt: 'json',
        q: query
      })
      .get()(function(err, res, body) {
        var video, videos;
        videos = JSON.parse(body);
        videos = videos.feed.entry;

        if (!videos) {
          return msg.send("Sorry, I couldn't find any \"{0}\" videos."
            .format(query));
        }

        video = videos[0];

        return video.link.forEach(function(link) {
          if (link.rel === "alternate" && link.type === "text/html") {
            return msg.send(link.href);
          }
        });
      });
    });
  };

}).call(this);
