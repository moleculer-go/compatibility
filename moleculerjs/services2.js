"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter, nodeID: process.env["NODE_ID"], logLevel: "trace"});

let looper = false;

broker.createService({
  name: "monitor",
  actions: {
    start(ctx) {
      const params = ctx.params;
      console.log("monitor.start action params: ", params);
      ctx.emit("monitor.started", params);
    },
  },
    
  events: {
    "user.*": params => {
      console.log("user.* events - params: ", params);
      
    },
     "profile.*": params => {
      console.log("profile.* events - params: ", params);
      
    },
    "account.*": params => {
      console.log(
        "profile.* events - params: ",
        params
      );
    }, 
    "notifier.*": params => {
      console.log(
        "profile.* events - params: ",
        params
      );
    }, 
  }
});

broker.start();

console.log("wait for profile service to be available!");
broker.waitForServices(["account", "profile"]).then(_ => {
  console.log("profile service is available!");
  broker.call("profile.create", { name: "John Doe" });
  broker.call("monitor.start", { services: ["account", "profile"] });
});
 
// // Time bomb
// setInterval(_ => {
//   process.exit();
// }, 5 * 60 * 1000); // 5 minutes
