"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter });

broker.createService({
  name: "profile",
  actions: {
    create(ctx) {
      const user = ctx.params;
      console.log("[moleculer-JS] profile.create action user: ", user);
      const { id, name, email } = user;
      const profile = { user: { id, name, email }, type: "web-user" };
      ctx.emit("profile.created", profile);
      return profile;
    },
    async finish() {
      console.log(
        "profile.finish called! will stop broker and finish process."
      );
      await broker.Stop();
      process.exit();
    }
  },
  events: {
    "user.created": user => {
      console.log("[moleculer-JS] user.created event! - user: ", user);
      broker.call("profile.create", user);
    }
  }
});
broker.createService({
  name: "account",
  events: {
    "profile.created": profile => {
      console.log(
        "[moleculer-JS] account service profile.created event! - profile: ",
        profile
      );
    }
  }
});

broker.start();
