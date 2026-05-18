import { TelegramClient } from "telegram";
import { StringSession } from "telegram/sessions/index.js";

const apiId = parseInt(Deno.env.get("TELEGRAM_API_ID") || "0");
const apiHash = Deno.env.get("TELEGRAM_API_HASH") || "";
const session = new StringSession("");

const client = new TelegramClient(session, apiId, apiHash, {
	connectionRetries: 5,
});

export async function fetchTelegramMessages(
	channelHandle: string,
): Promise<string[]> {
	try {
		await client.connect();

		const messages: string[] = [];
		const entity = await client.getEntity(channelHandle);

		const messageIter = client.iterMessages(entity, { limit: 70 });
		for await (const message of messageIter) {
			if (message.text) {
				messages.push(message.text);
			}
		}

		return messages;
	} catch (error) {
		console.error(`Could not fetch telegram channel ${channelHandle}`, error);
		return [];
	}
}
