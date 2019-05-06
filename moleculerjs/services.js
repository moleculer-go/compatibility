"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { ServiceBroker } = require("moleculer");

const broker = new ServiceBroker({ transporter });

broker.createService({
  name: "profile",
  actions: {
    create(ctx, params) {
      const { id, name, email } = params;
      const profile = { user: { id, name, email }, type: "web-user" };
      ctx.emit("profile.created", profile);
      return profile;
    },
    async finish() {
      ctx.emit("profile.finish called! will stop broker and finish process.");
      await broker.Stop();
      process.exit();
    }
  },
  events: {
    "user.created": user => {
      console.log("user.created event! - user: ", user);
      broker.call("profile.create", user);
    }
  }
});

broker.start();
