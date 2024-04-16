"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter, nodeID: process.env["NODE_ID"], logLevel: "trace"});

const monitorStore = {
  user:[],
  profile:[],
  account:[],
  notifier:[]
};

broker.createService({
  name: "monitor",
  actions: {
    start(ctx) {
      const params = ctx.params;
      console.log("monitor.start action params: ", params);
      ctx.emit("monitor.started", params);
    },

    allEvents(ctx) {
      return monitorStore;
    }
  },
    
  events: {
    "user.*": params => {
      console.log("user.* events - params: ", params);
      monitorStore.user.push(params);
    },
     "profile.*": params => {
      console.log("profile.* events - params: ", params);
      monitorStore.profile.push(params);
    },
    "account.*": params => {
      console.log(
        "profile.* events - params: ",
        params
      );
      monitorStore.account.push(params);
    }, 
    "notifier.*": params => {
      console.log(
        "profile.* events - params: ",
        params
      );
      monitorStore.notifier.push(params);
    }, 
  }
});

broker.start();

console.log("wait for monitor service to be available!");
broker.waitForServices(["monitor", "account", "profile", "user", "data"]).then(_ => {
  console.log("monitor service dependencies are available!"); 
  broker.call("monitor.start", { services: ["account", "profile", "user", "data"] });
});