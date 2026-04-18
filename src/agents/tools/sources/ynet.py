import logging

import feedparser

from agents.tools.sources.common.rss import RSSFeed

logger = logging.getLogger(__name__)


def get_ynet() -> RSSFeed:
	"""
	Fetch slightly left-leaning news outlet YNet.
	Political orientation: -3 (On a scale of [-10, -10])
	"""

	raw_feed = feedparser.parse("https://www.ynet.co.il/Integration/StoryRss2.xml")
	logger.debug(f"Fetched the news from YNet {raw_feed=}")

	return RSSFeed.model_validate(raw_feed)
