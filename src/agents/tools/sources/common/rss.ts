import { z } from "zod";

export const RSSFeedItemSchema = z.object({
	title: z.string(),
	link: z.string(),
	summary: z.string(),
	published: z.string(),
});

export type RSSFeedItem = z.infer<typeof RSSFeedItemSchema>;

export const RSSFeedSchema = z.object({
	entries: z.array(RSSFeedItemSchema),
});

export type RSSFeed = z.infer<typeof RSSFeedSchema>;
