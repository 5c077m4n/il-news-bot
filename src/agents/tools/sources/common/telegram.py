import logging
import os
from datetime import datetime, timedelta
from uuid import uuid4

from telethon import TelegramClient
from telethon.tl.custom.message import Message

from db import Article, MongoDB, Source

_TELEGRAM_SESSION = "/tmp/abu-ali-express-fetcher-bot"

client = TelegramClient(
	session=_TELEGRAM_SESSION,
	api_id=int(os.environ.get("TELEGRAM_API_ID", "0")),
	api_hash=os.environ.get("TELEGRAM_API_HASH", ""),
)
logger = logging.getLogger(__name__)


async def fetch_telegram_messages(channel_handle: str, source: Source) -> list[str]:
	try:
		with MongoDB() as db:
			latest_article = db.get_latest_article(source)
			if latest_article and (
				datetime.now() - latest_article.created_at.replace(tzinfo=None)
			) < timedelta(hours=1):
				return [a.description for a in db.get_articles(source)]

			async with client.takeout() as takeout:
				articles = []
				async for message in takeout.iter_messages(
					chat=channel_handle, wait_time=0, limit=70
				):
					if isinstance(message, Message) and message.text:
						articles.append(
							Article(
								id=uuid4(),
								title="Telegram Message",
								description=message.text,
								link="",
								source=source,
								created_at=message.date.replace(tzinfo=None)
								if message.date
								else datetime.now(),
							)
						)

					if articles:
						db.batch_create_articles(articles)

				return [a.description for a in articles]
	except Exception:
		logger.warning(
			f"Could not fetch telegram channel {channel_handle}",
			exc_info=True,
		)
		return []
