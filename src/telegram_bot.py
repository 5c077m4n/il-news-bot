import logging
import os

from telegram import Update
from telegram.constants import ParseMode
from telegram.ext import (
	Application,
	CommandHandler,
	ContextTypes,
	MessageHandler,
	filters,
)
from telegram.helpers import escape_markdown

from agents.graph import free_query, query
from agents.prompt_guard import DirtyInputException, DirtyOutputException

logging.getLogger("httpx").setLevel(logging.WARNING)
logger = logging.getLogger(__name__)


async def start(update: Update, _context: ContextTypes.DEFAULT_TYPE) -> None:
	user = update.effective_user

	if user and update.message:
		logger.info(f"{user.id=} has started")
		await update.message.reply_html(
			rf"Hi {user.mention_html()}! How can I help you today?"
		)


async def get_summary(
	update: Update,
	context: ContextTypes.DEFAULT_TYPE,
) -> None:
	if chat := update.effective_chat:
		if user := update.effective_user:
			logger.info(f"{user.id=} has requested a summary in {chat.id=}")

		response = await query()
		text = (
			escape_markdown(strip_resp)
			if response and (strip_resp := response.strip())
			else "Couldn't find anything..."
		)

		logger.info(f"{chat.id=} has responded with {text}")
		await context.bot.send_message(
			chat_id=chat.id,
			text=text,
			parse_mode=ParseMode.HTML,
		)


async def handle_free_text(
	update: Update,
	context: ContextTypes.DEFAULT_TYPE,
) -> None:
	if (
		update.message
		and (msg := update.message.text)
		and bool(chat := update.effective_chat)
	):
		if user := update.effective_user:
			logger.info(f"Recieved a message from {user.id=}: {msg} in {chat.id=}")

		response = await free_query(prompt=msg)
		text = (
			strip_resp
			if response and (strip_resp := response.strip())
			else "Couldn't find anything..."
		)
		logger.info(f"Responding to {chat.id=} with {text}")

		await context.bot.send_message(
			chat_id=chat.id,
			text=text,
			parse_mode=ParseMode.HTML,
		)


async def error_handler(update: object, context: ContextTypes.DEFAULT_TYPE) -> None:
	logger.error("Exception while handling an update:", exc_info=context.error)

	if (
		context.error
		and isinstance(update, Update)
		and bool(chat := update.effective_chat)
	):
		logger.error(f"An error has occured {context.error=} in {chat.id=}")

		if isinstance(context.error, DirtyInputException):
			await context.bot.send_message(
				chat_id=chat.id,
				text="Sorry, bad input",
				parse_mode=ParseMode.HTML,
			)
			return
		if isinstance(context.error, DirtyOutputException):
			await context.bot.send_message(
				chat_id=chat.id,
				text="Sorry, bad output",
				parse_mode=ParseMode.HTML,
			)
			return

		await context.bot.send_message(
			chat_id=chat.id,
			text="Sorry, my bad...",
			parse_mode=ParseMode.HTML,
		)


async def register_bot_commnads(app: Application) -> None:
	await app.bot.set_my_commands(
		[
			("summary", "Fetch the latest news"),
		]
	)


def run_telegram_bot() -> None:
	bot_token = os.getenv("TELEGRAM_BOT_TOKEN")
	if not bot_token:
		raise Exception("Telegram bot token was not found")

	app = (
		Application.builder().token(bot_token).post_init(register_bot_commnads).build()
	)
	app.add_handlers(
		[
			CommandHandler("start", start),
			CommandHandler("summary", get_summary),
			MessageHandler(filters.TEXT & ~filters.COMMAND, handle_free_text),
		]
	)
	app.add_error_handler(error_handler)

	app.run_polling(allowed_updates=Update.ALL_TYPES)
