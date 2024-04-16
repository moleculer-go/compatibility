"use strict";
 
const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter: "TCP", nodeID: `account-node-${Math.random() * 1000}`, logLevel: "info"});

let looper = false;

const accountUpates = [];
broker.createService({
  name: "account",
  actions: {
    async unregister(ctx) {
      console.log("account.unregister called! will un-register service.");
      await broker.destroyService("account");
      return "Service un-registered!";
    },
    async check(ctx) {
      console.log("account.check");
      const random = Math.floor(Math.random() * 100);
      ctx.emit("account.check", { random });
      return random;
    },
    async clear(ctx) {
      console.log("account.clear");
      ctx.emit("account.cleared",params);
      accountUpates.length = 0;
      return params;
    },
     async update(ctx) {
      console.log("account.update");
      const {params} = ctx;
      params.updated = new Date();
      ctx.emit("account.updated",params);
      accountUpates.push(params);
      return params;
    },
    async bulkUpdate(ctx) {
      const {params} = ctx;
      const {data, action} = params;
      console.log("account.bulkUpdate action: ", action, " data.length: ", data.length);

      //if action is "divide" it will return a list half the size of the input list (data)
      //the result list must contain a random set of items from the original list and cannot be duplicates
      let result = [];
      if (action === "divide") {
        const half = Math.floor(data.length / 2);
        for (let i = 0; i < half; i++) {
          let randomIndex = Math.floor(Math.random() * data.length);
          let randomItem = data[randomIndex];
          if (!result.includes(randomItem)) {
            result.push(await ctx.call("user.update", randomItem));
          }
        }
      } 
      //if action is "multiply" it will return a list double the size of the input list (data)
      //the result list must contain a random set of items from the original list. and a max of 2 duplicates is allowed ( to allow for the double size)
      else if (action === "multiply") {
        let count = {};
        for (let i = 0; i < data.length * 2; i++) {
          let randomIndex = Math.floor(Math.random() * data.length);
          let randomItem = data[randomIndex];
          if (!count[randomItem]) {
            count[randomItem] = 0;
          }
          if (count[randomItem] < 2) {
            result.push(await ctx.call("user.update", randomItem));
            count[randomItem]++;
          }
        }
      }

      return result;
    }
  },
  events: {
    "profile.created": profile => {
      console.log(
        "[moleculer-JS] account service profile.created event! - profile: ",
        profile
      );
    },
    "profile.loopevent": params => {
      console.log(
        "[moleculer-JS] account service profile.loopevent event! - params: ",
        params
      );
    },
    "notifier.sent": params => {
      console.log(
        "[moleculer-JS] account service notifier.sent event! - params: ",
        params
      );
    }
  }
});

broker.start();
 