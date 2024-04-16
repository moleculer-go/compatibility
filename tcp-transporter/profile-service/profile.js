"use strict";

const transporter = process.argv[2];
console.log("Start Moleculer JS with transporter: " + transporter);

const { set } = require("lodash");
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
      const { id, name, email, age } = user;
      const profile = { user: { id, name, email, age }, type: "web-user" };
      ctx.emit("profile.created", profile);

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

broker.start();
 
console.log("wait for profile service to be available!");
broker.waitForServices(["profile"]).then(_ => {
  console.log("profile service is available!");

  setInterval(_ => {
    const user = getRandomUser()
    broker.call("user.create", user);
  }, 5000);
});


function getRandomUser() {
    let name = names[Math.floor(Math.random() * names.length)];
    let domain = domains[Math.floor(Math.random() * domains.length)];
    let email = name.toLowerCase().replace(' ', '.') + "@" + domain;
    let obj = {
        name: name,
        email: email,
        age: Math.floor(Math.random() * 60)
    };
    return obj;
}

const names = ["John", "Jane", "Mary", "James", "Emma", "Jacob", "Olivia", "Noah", "Ava", "Liam", "Sophia", "Mason", "Isabella", 
                "William", "Mia", "Ethan", "Charlotte", "Michael", "Harper", "Alexander", "Emily", "Benjamin", "Abigail", "Daniel", 
                "Amelia", "David", "Evelyn", "Joseph", "Elizabeth", "Samuel", "Sofia", "Matthew", "Avery", "Lucas", "Ella", "Jackson", 
                "Scarlett", "Bryson", "Grace", "Carter", "Chloe", "Jayden", "Victoria", "Henry", "Madison", "Michael", "Luna", "Daniel", 
                "Mila", "Elijah", "Hannah", "Benjamin", "Lily", "Matthew", "Addison", "Joseph", "Nora", "Samuel", "Zoe", "David", "Stella", 
                "Carter", "Natalie", "Owen", "Emilia", "Ryan", "Everly", "Luke", "Leah", "Christian", "Aubrey", "Hunter", "Ellie", "Jaxon", 
                "Zoey", "Nolan", "Savannah", "Easton", "Brooklyn", "Nicholas", "Bella", "Ezra", "Paisley", "Colton", "Skylar", "Adam", "Alan", 
                "Albert", "Alexander", "Alfred", "Andrew", "Anthony", "Arthur", "Benjamin", "Bernard", "Charles", "Christopher", "Daniel", "David", 
                "Dennis", "Donald", "Edward", "Eric", "Francis", "Frank", "Frederick", "George", "Gerald", "Gregory", "Harold", "Harry", "Henry", "Howard", 
                "Jack", "James", "Jason", "Jeffrey", "Jeremy", "Jerry", "John", "Jonathan", "Joseph", "Justin", "Keith", "Kenneth", "Kevin", "Larry", 
                "Lawrence", "Louis", "Mark", "Martin", "Matthew", "Michael", "Nathan", "Nicholas", "Patrick", "Paul", "Peter", "Philip", "Raymond", 
                "Richard", "Robert", "Roger", "Ronald", "Russell", "Samuel", "Scott", "Sean", "Stephen", "Steven", "Thomas", "Timothy", "Walter", 
                "William", "Winston", "Zachary", "Abigail", "Alexandra", "Alice", "Amanda", "Amelia", "Amy", "Andrea", "Angela", "Ann", "Anna", 
                "Anne", "Audrey", "Ava", "Barbara", "Beatrice", "Bella", "Bernadette", "Bertha", "Bessie", "Beth", "Betty", "Beverly", "Bonnie", 
                "Brenda", "Bridget", "Brittany", "Camilla", "Candice", "Carla", "Carmen", "Carol", "Caroline", "Catherine", "Cecilia", "Celeste", 
                "Charlotte", "Cheryl", "Christina", "Christine", "Cindy", "Claire", "Clara", "Claudia", "Colleen", "Constance", "Cora", "Courtney", 
                "Crystal", "Cynthia", "Daisy", "Dana", "Danielle", "Daphne", "Darlene", "Dawn", "Deborah", "Debra", "Denise", "Diana", "Diane", "Donna", 
                "Dora", "Doris", "Dorothy", "Edith", "Edna", "Eileen", "Elaine", "Eleanor", "Elizabeth", "Ellen", "Emily", "Emma", "Erica", "Erin", "Esther", 
                "Ethel", "Eva", "Evelyn", "Faith", "Faye", "Felicia", "Fiona", "Florence", "Frances", "Gabrielle", "Gail", "Gemma", "Genevieve", "Georgia", 
                "Geraldine", "Gillian", "Gina", "Gloria", "Grace", "Gwendolyn", "Hannah", "Harriet", "Hazel", "Heather", "Helen", "Hilary", "Holly", 
                "Ida", "Irene", "Iris", "Isabel", "Isabella", "Isabelle", "Ivy", "Jacqueline", "Jane", "Janet", "Janice", "Jean", "Jeanette", "Jeanne", 
                "Jenna", "Jennifer", "Jessica", "Jill", "Joan", "Joanna", "Jocelyn", "Jodi", "Jodie", "Johanna", "Josephine", "Joy", "Joyce", "Judith", "Judy", 
                "Julia", "Julie", "June", "Karen", "Katherine", "Kathleen", "Kathryn", "Katie", "Kay", "Kendra", "Kim", "Kimberly", "Kristen", "Kristin", 
                "Laura", "Lauren", "Leah", "Leslie", "Lillian", "Linda", "Lisa", "Lois", "Loretta", "Lori", "Louise", "Lucy", "Lydia", "Lynn", "Mabel", "Madeline",
                 "Maggie", "Marcia", "Margaret", "Maria", "Marian", "Marianne", "Marie", "Marilyn", "Marion", "Marjorie", "Martha", "Mary", "Maureen", "Megan", 
                 "Melanie", "Melinda", "Melissa", "Meredith", "Michele", "Michelle", "Mildred", "Miranda", "Molly", "Monica", "Nancy", "Naomi", "Natalie", "Natasha", 
                 "Nellie", "Nicola", "Nicole", "Nina", "Norma", "Olivia", "Pamela", "Patricia", "Paula", "Pauline", "Peggy", "Penelope", "Phyllis", "Priscilla", "Rachel", 
                 "Rebecca", "Renee", "Rhonda", "Rita", "Roberta", "Robin", "Rosalind", "Rose", "Rosemary", "Ruby", "Ruth", "Sally", "Samantha", "Sandra", "Sara", "Sarah", 
                 "Savannah", "Selma", "Sharon", "Sheila", "Shelley", "Sherry", "Shirley", "Sonia", "Stella", "Stephanie", "Susan", "Suzanne", "Sylvia", "Tanya", "Tara", 
                 "Teresa", "Theresa", "Tina", "Tracy", "Ursula", "Valerie", "Vanessa", "Vera", "Veronica", "Victoria", "Violet", "Virginia", "Vivian", "Wanda", "Wendy", "Wilma", "Yvonne", "Zoe"];
const domains = ["example.com", "test.com", "sample.com", "demo.com", "trial.com", "mydomain.com", "email.com", "mailbox.com", "post.com", "letter.com"];

 