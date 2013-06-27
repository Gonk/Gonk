(function() {
  module.exports = function(robot) { 
    robot.respond(/rimshot/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=rimshot&play=true");
    });

    robot.respond(/(csi|caruso)?/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=csi&play=true");
    });

    robot.respond(/cowbell/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=cowbell&play=true");
    });

    robot.respond(/crickets/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=crickets&play=true");
    });

    robot.respond(/downer/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=downer&play=true");
    });

    robot.respond(/drum( )?roll/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=drumroll&play=true");
    });

    robot.respond(/gobble/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=gobble&play=true");
    });

    robot.respond(/gong/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=gong&play=true");
    });

    robot.respond(/price( )?is( )?wrong/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=priceiswrong&play=true");
    });

    robot.respond(/reveille/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=reveille&play=true");
    });

    robot.respond(/reading( )?rainbow/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=readingrainbow&play=true");
    });

    robot.respond(/slide( )?whistle/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=slidewhistle&play=true");
    });

    robot.respond(/ye+( )?ha+(w+)?/i, function(msg) {
      return msg.send("http://instantrimshot.com/index.php?sound=yeehaw&play=true");
    });
  }
}).call(this);
