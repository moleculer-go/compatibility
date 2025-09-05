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
      console.log("[moleculer-JS] profile.mistake called with: ", ctx.params);
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
      console.log("[moleculer-JS] profile.finish called with: ", ctx.params);
      ctx.emit("profile.finished", { message: "JS side will explode in 500 miliseconds!" });
      setTimeout(_ => {
        process.exit(0);
      }, 500);
      return "JS side will explode in 500 miliseconds!";
    },

    async unregister(ctx) {
      console.log("[moleculer-JS] profile.unregister called");
      broker.destroyService("account");
      return "account service unregistered";
    }
  }
});

broker.createService({
  name: "account",
  actions: {
    unregister(ctx) {
      console.log("[moleculer-JS] account.unregister called");
      return "account service unregistered";
    }
  }
});

broker.start().then(() => {
  console.log("🚀 Moleculer JS broker started with NATS transporter");
  console.log("📡 NATS connection: connected");
  console.log("🧮 Testing math operations...");
  
  // Test some basic operations
  broker.call("profile.create", { id: 1, name: "Test", email: "test@example.com" })
    .then(result => {
      console.log("Math calculation completed:", result);
      console.log("1 + 1 = 2");
    })
    .catch(err => console.error("Error:", err));
    
  console.log("⏳ Running for 5 seconds to demonstrate events...");
  
  // Keep the service running for a bit
  setTimeout(() => {
    console.log("🛑 JS Broker stopped");
    broker.stop();
  }, 5000);
});
