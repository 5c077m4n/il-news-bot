import { HumanMessage, SystemMessage } from "@langchain/core/messages";
import { END, type GraphNode, START, StateGraph } from "@langchain/langgraph";
import { ChatOpenRouter as OpenRouter } from "@langchain/openrouter";
import { z } from "zod";
import { getAbuAliExpress } from "./tools/sources/abu_ali_express.ts";
import { getIsraelHayom } from "./tools/sources/israel_hayom.ts";
import { getMarker } from "./tools/sources/marker.ts";
import { getYnet } from "./tools/sources/ynet.ts";

const ArticleSchema = z.object({
	title: z.string(),
	content: z.string(),
	links: z.array(z.string()),
});
const AnchorResponseSchema = z.object({ articles: z.array(ArticleSchema) });

const StateSchema = z.object({
	prompt: z.string(),
	leftNews: AnchorResponseSchema.optional(),
	rightNews: AnchorResponseSchema.optional(),
	allNews: AnchorResponseSchema.optional(),
});
type State = z.infer<typeof StateSchema>;

const llm = new OpenRouter({
	model: "google/gemini-3.1-flash-lite-preview",
	apiKey: Deno.env.get("OPENROUTER_API_KEY"),
	baseURL: "https://openrouter.ai/api/v1",
});
const newsAnchorLlm = llm.withStructuredOutput(AnchorResponseSchema);

const callLeftyAnchor: GraphNode<State> = async (state) => {
	const ynet = await getYnet();
	const marker = await getMarker();

	const messages = [
		new SystemMessage(
			"Act as a progressive news anchor... (skipped some context for brevity)",
		),
		new SystemMessage(`YNet feed: ${JSON.stringify(ynet)}`),
		new SystemMessage(`Marker feed: ${JSON.stringify(marker)}`),
		new HumanMessage(state.prompt || "Please get me all the current news"),
	];

	const response = await newsAnchorLlm.invoke(messages);
	return { leftNews: response };
};

const callRightyAnchor: GraphNode<State> = async (state) => {
	const israelHayom = await getIsraelHayom();
	const abuAliExpress = await getAbuAliExpress();

	const messages = [
		new SystemMessage("Act as a center-right news anchor..."),
		new SystemMessage(`Israel Hayom feed: ${JSON.stringify(israelHayom)}`),
		new SystemMessage(`Abu Ali Express feed: ${JSON.stringify(abuAliExpress)}`),
		new HumanMessage(state.prompt || "Please get me all the current news"),
	];

	const response = await newsAnchorLlm.invoke(messages);
	return { rightNews: response };
};

const aggregator: GraphNode<State> = async (state) => {
	const messages = [
		new SystemMessage("You are a fact checker..."),
		new SystemMessage(`Left news: ${JSON.stringify(state.leftNews)}`),
		new SystemMessage(`Right news: ${JSON.stringify(state.rightNews)}`),
	];

	const response = await newsAnchorLlm.invoke(messages);
	return { allNews: response };
};

const graph = new StateGraph(StateSchema)
	.addNode("lefty", callLeftyAnchor)
	.addNode("righty", callRightyAnchor)
	.addNode("aggregator", aggregator)
	.addEdge(START, "lefty")
	.addEdge(START, "righty")
	.addEdge("lefty", "aggregator")
	.addEdge("righty", "aggregator")
	.addEdge("aggregator", END);
export const workflow = graph.compile();

export async function query(
	prompt = "Please get me all the current news",
): Promise<string> {
	const result = await workflow.invoke({ prompt });
	return JSON.stringify(result.allNews);
}
