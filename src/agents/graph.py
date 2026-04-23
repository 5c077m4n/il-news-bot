import logging
import operator
from datetime import datetime
from typing import Annotated, Literal

from langchain.messages import HumanMessage, SystemMessage
from langchain_openrouter import ChatOpenRouter
from langgraph.graph import END, START, StateGraph
from pydantic import BaseModel, Field

from agents.prompt_guard import clean_input, clean_ouput
from agents.tools.sources.israel_hayom import get_israel_hayom
from agents.tools.sources.marker import get_marker
from agents.tools.sources.ynet import get_ynet

logger = logging.getLogger(__name__)


class Article(BaseModel):
	title: str
	content: str
	links: Annotated[
		list[str],
		Field(default_factory=list, description="List of sources"),
	]

	def __str__(self) -> str:
		return f"""<strong>{self.title}</strong>
{self.content}
{" | ".join(f'<a href="{link}">{i + 1}</a>' for i, link in enumerate(self.links))}
"""


class AnchorResponse(BaseModel):
	articles: list[Article]


llm = ChatOpenRouter(model="google/gemini-3.1-flash-lite-preview")
news_anchor_llm = llm.with_structured_output(AnchorResponse)


class State(BaseModel):
	prompt: str | None = None
	left_news_items: Annotated[list[Article] | None, operator.add] = None
	right_news_items: Annotated[list[Article] | None, operator.add] = None
	all_news_items: Annotated[list[Article] | None, operator.add] = None


async def call_lefty_anchor(
	state: State,
) -> dict[Literal["left_news_items"], list[Article]]:
	messages: list[SystemMessage | HumanMessage] = [
		SystemMessage(
			content="""
				Act as a progressive news anchor who is principled, calm,
				and meticulous.
				Your perspective leans left—prioritizing social justice,
				environmental protection,
				and economic equality—but your primary allegiance is to the truth.
				Every headline you deliver must be accompanied by a specific,
				credible source.
				Avoid hyperbole; let the data and the ethics of the story drive the
				narrative. Your tone is professional, empathetic, and intellectually
				rigorous.
				"""
		),
		SystemMessage(
			content="""
				**Do not** send a headline without at least one link to the original
				source (the more souces the better).
				"""
		),
		SystemMessage(content=f"The time right now is {datetime.now()}"),
		SystemMessage(content=f"YNet feed: {get_ynet()}"),
		SystemMessage(content=f"The marker feed: {get_marker()}"),
		SystemMessage(
			content="Keep your responses short and to the point - no fluff words."
		),
		HumanMessage(content=state.prompt or "Please get me all the current news"),
	]

	response = await news_anchor_llm.ainvoke(messages)
	anchor_resopnse = AnchorResponse.model_validate(response)

	return {"left_news_items": anchor_resopnse.articles}


async def call_righty_anchor(
	state: State,
) -> dict[Literal["right_news_items"], list[Article]]:
	messages: list[SystemMessage | HumanMessage] = [
		SystemMessage(
			content="""
				Act as a principled, center-right news anchor.
				Your tone is professional, traditional, and analytical.
				You prioritize individual liberty, fiscal responsibility, and local
				governance. Crucially, every headline must be followed by a specific,
				credible source or data point. Avoid hyperbole; focus on interpreting
				current events through a conservative lens while maintaining strict
				journalistic integrity and factual accuracy.
				"""
		),
		SystemMessage(
			content="""
				**Do not** send a headline without at least one link to the original
				source (the more souces the better).
				"""
		),
		SystemMessage(content=f"The time right now is {datetime.now()}"),
		SystemMessage(content=f"Israel Hayom feed: {get_israel_hayom()}"),
		# SystemMessage(content=f"Abu Ali Express feed: {await get_abu_ali_express()}"),
		SystemMessage(
			content="Keep your responses short and to the point - no fluff words."
		),
		HumanMessage(content=state.prompt or "Please get me all the current news"),
	]

	response = await news_anchor_llm.ainvoke(messages)
	anchor_response = AnchorResponse.model_validate(response)

	return {"right_news_items": anchor_response.articles}


async def aggregator(state: State) -> dict[Literal["all_news_items"], list[Article]]:
	messages: list[SystemMessage | HumanMessage] = [
		SystemMessage(
			content="""
				You are a fact checker that makes sure that any and all information
				passed through you is true, you make sure that all links are valid and
				return a non-error status code (2**) when opening, that stories are
				mentioned more than once (a good indication but not definitive).
				After validating all the news lists then you'll return only one that
				includes all good items from both of them without duplications (if a
				story is in more than one article then just attach all relevant links).
				"""
		),
		SystemMessage(
			content="""
			Make sure that your output is **valid** telegram supported HTML,
			using only the tags: <b>, <strong>, <i>, <em>, <u>, <ins>, <s>, <strike>,
			<code>, <pre>, and <a>, and "\n" for line breaks
			"""
		),
		SystemMessage(
			content="""
			If there are no news articles then just say that an error has occured in
			fetching the news and to try again later
			"""
		),
		SystemMessage(content=f"The time right now is {datetime.now()}"),
		SystemMessage(
			content=f"""
			These are the left leaning items:
			{
				"\n".join(
					i.model_dump().__str__()
					for i in (state.left_news_items or [])[0:10]
				)
			}
			"""
		),
		SystemMessage(
			content=f"""
			These are the right leaning items:
			{
				"\n".join(
					i.model_dump().__str__()
					for i in (state.right_news_items or [])[0:10]
				)
			}
			"""
		),
	]

	response = await news_anchor_llm.ainvoke(messages)
	anchor_response = AnchorResponse.model_validate(response)

	return {"all_news_items": anchor_response.articles}


state_graph = StateGraph(State)
state_graph.add_node(node=call_lefty_anchor.__name__, action=call_lefty_anchor)
state_graph.add_node(node=call_righty_anchor.__name__, action=call_righty_anchor)
state_graph.add_node(node=aggregator.__name__, action=aggregator)

state_graph.add_edge(start_key=START, end_key=call_lefty_anchor.__name__)
state_graph.add_edge(start_key=START, end_key=call_righty_anchor.__name__)
state_graph.add_edge(start_key=call_lefty_anchor.__name__, end_key=aggregator.__name__)
state_graph.add_edge(start_key=call_righty_anchor.__name__, end_key=aggregator.__name__)
state_graph.add_edge(start_key=aggregator.__name__, end_key=END)

workflow = state_graph.compile()


async def query() -> str:
	response = await workflow.ainvoke(State())
	news_items = response.get("all_news_items", [])
	return "\n".join(n.__str__() for n in news_items)


async def free_query(prompt: str) -> str:
	clean_prompt = clean_input(prompt)
	response = await workflow.ainvoke(State(prompt=clean_prompt))
	news_items = response.get("all_news_items", [])

	news = clean_ouput(
		response_text="\n".join(n.__str__() for n in news_items),
		sanitized_prompt=clean_prompt,
	)
	return news
