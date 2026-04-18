import logging

import feedparser

from agents.tools.sources.common.rss import RSSFeed

logger = logging.getLogger(__name__)


def get_israel_hayom() -> RSSFeed:
	"""
	Fetch slightly right leaning news outlet Israel Hayom.
	Political orientation: 3 (On a scale of [-10, -10])
	"""

	raw_feed = feedparser.parse("https://www.israelhayom.co.il/rss.xml")
	logger.debug(f"Fetched the news from Israel Hayom {raw_feed=}")

	return RSSFeed.model_validate(raw_feed)
