import { load } from "@std/dotenv";
import { setupBot } from "./telegram/bot.ts";

async function main(): Promise<void> {
	await load({ export: true });
	setupBot();
}

main().catch((e) => {
	console.error(e);
	Deno.exit(1);
});
