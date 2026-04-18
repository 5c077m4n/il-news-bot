import logging
import os

from dotenv import load_dotenv

from telegram_bot import run_telegram_bot

logging.basicConfig(
	format="[%(asctime)s] %(levelname)s - %(name)s - %(message)s",
	level=os.environ.get("LOG_LEVEL", logging.INFO),
)


def main() -> None:
	load_dotenv(verbose=True)
	run_telegram_bot()


if __name__ == "__main__":
	main()
