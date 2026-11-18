// /guides/security.md and friends: each docs page as plain markdown.
import type { APIRoute, GetStaticPaths } from 'astro';
import { getSections, MARKDOWN, toMarkdown } from '../llms';

export const getStaticPaths: GetStaticPaths = async () => {
	const sections = await getSections();
	return sections
		.flatMap((section) => section.docs)
		.map((doc) => ({ params: { slug: doc.id }, props: { doc } }));
};

export const GET: APIRoute = ({ props }) =>
	new Response(toMarkdown(props.doc) + '\n', { headers: MARKDOWN });
