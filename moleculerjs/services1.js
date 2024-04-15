"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter, nodeID: process.env["NODE_ID"], logLevel: "trace"});

let looper = false;

broker.createService({
  name: "profile",
  actions: {

    listServices(ctx) {
      return ctx.call("$node.services");
    },

    create(ctx) {
      const user = ctx.params;
      console.log("[moleculer-JS] profile.create action user: ", user);
      const { id, name, email } = user;
      const profile = { user: { id, name, email }, type: "web-user" };
      ctx.emit("profile.created", profile);

      setInterval(_ => {
        if (looper) {
          ctx.broadcast("profile.loopevent", { name: "loop" });
        }
      }, 2000);

      return profile;
    },

    async metarepeat(ctx) {
      console.log("[moleculer-JS] profile.metarepeat ctx.meta: ", ctx.meta);
      return { meta: ctx.meta, params: ctx.params };
    },

    async mistake(ctx) {
      let panixError = "";
      let failError = "";
      await broker.waitForServices("user");
      try {
        const panic = await ctx.call(
          "user.panix",
          {},
          { meta: { name: "John", sword: "Valyrian Steel" } }
        );
      } catch (e) {
        console.log("error calling panic: ", e.message);
        panixError = e.message;
      }

      try {
        const fail = await ctx.call("user.fail", {});
      } catch (e) {
        failError = e.message;
        console.log("error calling fail: ", failError);
      }

      throw new Error(
        `Error from JS side! panixError: [${panixError}] failError: [${failError}]`
      );
    },

    async finish(ctx) {
      console.log(
        "profile.finish called! will stop broker and finish process."
      );
      broker.waitForServices("notifier");
      for (let index = 0; index < 5; index++) {
        const notification = await ctx.call("notifier.send", {
          title: "shutdown",
          index
        });
        console.log("profile.finish notification: ", notification);
      }
      looper = false;

      console.log(
        "profile.finish Notifications sent! will auto explode now..."
      );

      setTimeout(async () => {
        await broker.stop();
        process.exit();
      }, 1000);
      return "JS side will explode in 500 miliseconds!";
    },

    "mutationExample": {
      params: {
        name: { "type": "string", "optional": false },
        lastname: { "type": "string", "optional": true },
      },
      output: {
        eventId: "number",
        createdAt: "number"
      },
      graphql: "mutation",
      handler: async (ctx) => {
        console.log("do nothing...");
      },
    },

    "queryExample": {
      params: {
        name: { "type": "string", "optional": false },
        lastname: { "type": "string", "optional": true },
      },
      output: {
        eventId: "number",
        createdAt: "number"
      },
      graphql: "query",
      handler: async (ctx) => {
        console.log("do nothing...");
      },
    },
    async check(ctx) {
      console.log("profile.check");
      const random = Math.floor(Math.random() * 100);
      ctx.emit("profile.check", { random });
      return random;
    },
  },
  events: {
    "user.created": user => {
      console.log("[moleculer-JS] user.created event! - user: ", user);
      console.log("wait for user service to be available!");
      broker.waitForServices("user");

      broker.call("profile.create", user);
      broker.call("user.get", user);
    }
  }
});

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

// //un-register service
// setInterval(_ => {
   
// }, 5 * 1000); // 5 seconds



// // Time bomb
// setInterval(_ => {
//   process.exit();
// }, 5 * 60 * 1000); // 5 minutes
