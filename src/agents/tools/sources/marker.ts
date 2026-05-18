import Parser from "rss-parser";
import type { RSSFeed } from "./common/rss.ts";

const parser = new Parser();

export async function getMarker(): Promise<RSSFeed> {
	const feed = await parser.parseURL("https://www.themarker.com/srv/tm-news");
	return {
		entries: feed.items.map((item) => ({
			title: item.title || "",
			link: item.link || "",
			summary: item.contentSnippet || item.content || "",
			published: item.pubDate || "",
		})),
	};
}
