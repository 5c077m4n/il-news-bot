import Parser from "rss-parser";
import type { RSSFeed } from "./common/rss.ts";

const parser = new Parser();

export async function getIsraelHayom(): Promise<RSSFeed> {
	const feed = await parser.parseURL("https://www.israelhayom.co.il/rss.xml");

	return {
		entries: feed.items.map((item) => ({
			title: item.title || "",
			link: item.link || "",
			summary: item.contentSnippet || item.content || "",
			published: item.pubDate || "",
		})),
	};
}
