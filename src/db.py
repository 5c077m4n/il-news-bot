import logging
import os
from contextlib import AbstractContextManager
from datetime import datetime
from enum import Enum
from typing import Annotated
from uuid import UUID, uuid4

from pydantic import BaseModel, Field
from pymongo import MongoClient
from pymongo.collection import Collection
from pymongo.database import Database

logger = logging.getLogger(__name__)


class Source(str, Enum):
	AbuAliExpress = "AbuAliExpress"


class Article(BaseModel):
	id: Annotated[UUID, Field(default_factory=uuid4, alias="_id")]
	title: str
	description: str
	link: str
	source: Source
	created_at: Annotated[datetime, Field(default_factory=datetime.now)]


class MongoDB(AbstractContextManager):
	_client: MongoClient
	_database: Database
	_articles_collection: Collection

	def __init__(self):
		username = os.getenv("MONGODB_USERNAME")
		password = os.getenv("MONGODB_PASSWORD")
		database = os.getenv("MONGODB_DATABASE")
		host = os.getenv("MONGODB_HOST")
		if not (username or password or database or host):
			raise Exception("Couldn't find MongoDB connection data")

		self._client = MongoClient(
			host=f"mongodb://{username}:{password}@{host}/?retryWrites=true&w=majority"
		)
		self._database = self._client.get_database(database)
		self._articles_collection = self._database.get_collection("articles")

	def __exit__(self, _exc_type, _exc_value, _traceback) -> None:
		self._client.close()

	def get_articles(self, source: Source) -> list[Article]:
		with self._articles_collection.find(
			filter={"source": source},
			limit=50,
		) as cursor:
			return [Article.model_validate(item) for item in cursor]

	def get_latest_article(self, source: Source) -> Article | None:
		if article := self._articles_collection.find_one(
			filter={"source": source},
			sort=[("created_at", -1)],
		):
			return Article.model_validate(article)

	def batch_create_articles(self, articles: list[Article]) -> None:
		result = self._articles_collection.insert_many(a.model_dump() for a in articles)
		if len(articles) != len(result.inserted_ids):
			logger.warning(f"Not all articles where inserted into the DB; {articles=}")
