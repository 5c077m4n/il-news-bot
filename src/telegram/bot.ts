import { Bot } from "grammy";
import { query } from "../agents/graph.ts";

const botToken = Deno.env.get("TELEGRAM_BOT_TOKEN");
if (!botToken) {
	throw new Error("TELEGRAM_BOT_TOKEN is not defined");
}

const bot = new Bot(botToken);

export function setupBot() {
	bot.command("start", (ctx) => ctx.reply("Hi! How can I help you today?"));
	bot.command("summary", async (ctx) => {
		await ctx.reply("Fetching latest news...");
		try {
			const summary = await query();
			await ctx.reply(summary, { parse_mode: "HTML" });
		} catch (error) {
			console.error(error);
			await ctx.reply("Failed to fetch news.");
		}
	});

	bot.on("message:text", async (ctx) => {
		const { text } = ctx.message;
		await ctx.reply(`You said: ${text}. Processing...`);
	});

	bot.start();
}
