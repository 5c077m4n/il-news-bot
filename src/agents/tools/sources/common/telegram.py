import logging
import os

from telethon import TelegramClient

_TELEGRAM_SESSION = "/tmp/abu-ali-express-fetcher-bot"


logger = logging.getLogger(__name__)

bot = TelegramClient(
	session=_TELEGRAM_SESSION,
	api_id=int(os.environ.get("TELEGRAM_API_ID", "0")),
	api_hash=os.environ.get("TELEGRAM_API_HASH", ""),
).start(bot_token=os.environ.get("TELEGRAM_BOT_TOKEN", ""))


async def fetch_telegram_messages(channel_handle: str) -> list[str]:
	async with bot:
		try:
			messages = await bot.get_messages(channel_handle, limit=20)
			return [
				m.__dict__.get("text", "")
				for m in (messages if isinstance(messages, list) else [messages])
			]
		except Exception:
			logger.warning(
				f"Could not fetch telegram channel {channel_handle}",
				exc_info=True,
			)
			return []
