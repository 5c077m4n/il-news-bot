from pydantic import BaseModel


class RSSFeedItem(BaseModel):
	title: str
	link: str
	summary: str
	published: str


class RSSFeed(BaseModel):
	entries: list[RSSFeedItem]
