"use strict";
 
const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter: "TCP", nodeID: `monitor-node-${Math.random() * 1000}`, logLevel: "info"});

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
      broker.logger.info("monitor.start action params: ", params);
      ctx.emit("monitor.started", params);
    },

    allEvents(ctx) {
      return monitorStore;
    }
  },
    
  events: {
    "user.*": params => {
      broker.logger.info("user.* events - params: ", params);
      monitorStore.user.push(params);
    },
     "profile.*": params => {
      broker.logger.info("profile.* events - params: ", params);
      monitorStore.profile.push(params);
    },
    "account.*": params => {
      broker.logger.info(
        "profile.* events - params: ",
        params
      );
      monitorStore.account.push(params);
    }, 
    "notifier.*": params => {
      broker.logger.info(
        "profile.* events - params: ",
        params
      );
      monitorStore.notifier.push(params);
    }, 
  }
});

broker.start();

broker.logger.info("wait for monitor service to be available!");
broker.waitForServices(["monitor", "account", "profile", "user", "data"]).then(_ => {
  broker.logger.info("monitor service dependencies are available!"); 
  broker.call("monitor.start", { services: ["account", "profile", "user", "data"] });
});