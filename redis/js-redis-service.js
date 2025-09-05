const { ServiceBroker } = require("moleculer");

// Get Redis connection details from environment
const redisHost = process.env.REDIS_HOST || "localhost";
const redisPort = process.env.REDIS_PORT || "6379";
const redisUrl = `redis://${redisHost}:${redisPort}/2`; // Use DB 2 for testing

console.log(`Starting Moleculer JS with Redis transporter: ${redisUrl}`);

// Create broker with Redis transporter
const broker = new ServiceBroker({
    nodeID: "js-redis-node",
    logger: console,
    logLevel: "info",
    transporter: redisUrl
});

// Math service
broker.createService({
    name: "math",
    actions: {
        add: {
            params: {
                a: "number",
                b: "number"
            },
            handler(ctx) {
                const result = ctx.params.a + ctx.params.b;
                console.log(`Math.add: ${ctx.params.a} + ${ctx.params.b} = ${result}`);
                return { result };
            }
        },
        multiply: {
            params: {
                a: "number",
                b: "number"
            },
            handler(ctx) {
                const result = ctx.params.a * ctx.params.b;
                console.log(`Math.multiply: ${ctx.params.a} * ${ctx.params.b} = ${result}`);
                return { result };
            }
        }
    },
    events: {
        "math.calculated": {
            handler(ctx) {
                console.log("Math calculation completed:", ctx.params);
            }
        }
    }
});

// Calculator service that calls math service
broker.createService({
    name: "calculator",
    actions: {
        calculate: {
            params: {
                operation: "string",
                a: "number",
                b: "number"
            },
            async handler(ctx) {
                const { operation, a, b } = ctx.params;
                let result;

                switch (operation) {
                    case "add":
                        const addResult = await ctx.call("math.add", { a, b });
                        result = addResult.result;
                        break;
                    case "multiply":
                        const multiplyResult = await ctx.call("math.multiply", { a, b });
                        result = multiplyResult.result;
                        break;
                    default:
                        return { error: "Unknown operation" };
                }

                // Emit event
                ctx.emit("math.calculated", {
                    operation,
                    a,
                    b,
                    result
                });

                return { result };
            }
        }
    }
});

// Start the broker
broker.start().then(() => {
    console.log("🚀 Moleculer JS broker started with Redis transporter");
    console.log("📡 Redis connection: connected");
    
    // Test the services
    console.log("\n🧮 Testing math operations...");
    
    // Test addition
    broker.call("calculator.calculate", {
        operation: "add",
        a: 10,
        b: 5
    }).then(result => {
        console.log(`10 + 5 = ${result.result}`);
        
        // Test multiplication
        return broker.call("calculator.calculate", {
            operation: "multiply",
            a: 10,
            b: 5
        });
    }).then(result => {
        console.log(`10 * 5 = ${result.result}`);
        
        // Test direct math service
        return broker.call("math.add", {
            a: 20,
            b: 30
        });
    }).then(result => {
        console.log(`20 + 30 = ${result.result}`);
        
        // Keep running for a bit to demonstrate events
        console.log("\n⏳ Running for 5 seconds to demonstrate events...");
        setTimeout(() => {
            broker.stop();
            console.log("🛑 JS Broker stopped");
        }, 5000);
    }).catch(err => {
        console.error("Error:", err);
        broker.stop();
    });
}).catch(err => {
    console.error("Failed to start broker:", err);
    process.exit(1);
});
