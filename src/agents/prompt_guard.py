import logging

from llm_guard import scan_output, scan_prompt
from llm_guard.input_scanners import BanCode, InvisibleText, PromptInjection, TokenLimit
from llm_guard.output_scanners import MaliciousURLs, NoRefusal, Relevance


class DirtyInputException(Exception):
	message: str | None

	def __init__(self, message: str | None = None):
		super().__init__(message)
		self.message = message


class DirtyOutputException(Exception):
	message: str | None

	def __init__(self, message: str | None = None):
		super().__init__(message)
		self.message = message


logger = logging.getLogger(__name__)

input_scanners = [TokenLimit(), PromptInjection(), InvisibleText(), BanCode()]
output_scanners = [NoRefusal(), Relevance(), MaliciousURLs(threshold=0.8)]


def clean_input(prompt: str) -> str:
	sanitized_prompt, results_valid, results_score = scan_prompt(
		list(input_scanners),
		prompt,
	)
	logger.debug(f"Cleaned {prompt=} into {sanitized_prompt=}")

	if any(not result for result in results_valid.values()):
		logger.error(f"{prompt=} is not valid, {results_score=}")
		raise DirtyInputException()

	return sanitized_prompt


def clean_ouput(response_text: str, sanitized_prompt: str) -> str:
	sanitized_response_text, results_valid, results_score = scan_output(
		list(output_scanners),
		sanitized_prompt,
		response_text,
	)
	logger.debug(f"Cleaned {response_text=} into {sanitized_response_text=}")

	if any(not result for result in results_valid.values()):
		logger.error(f"{response_text=} is not valid, {results_score=}")
		raise DirtyOutputException()

	return sanitized_response_text
