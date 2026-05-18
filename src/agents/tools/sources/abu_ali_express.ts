import { fetchTelegramMessages } from "./common/telegram.ts";

export async function getAbuAliExpress(): Promise<string[]> {
	// Fetch from the handle
	return await fetchTelegramMessages("@abualiexpress");
}
