import logging

from agents.tools.sources.common.telegram import fetch_telegram_messages
from db import Source

logger = logging.getLogger(__name__)


async def get_abu_ali_express() -> list[str]:
	"""
	Fetch right-leaning news outlet Abu Ali Express
	Political orientation: 7 (On a scale of [-10, -10])
	"""
	messages = await fetch_telegram_messages(
		channel_handle="@abualiexpress",
		source=Source.AbuAliExpress,
	)
	logger.debug(f"Fetched {messages=} from Abu Ali Express")

	return messages
