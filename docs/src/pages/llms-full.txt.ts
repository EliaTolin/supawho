import type { APIRoute } from 'astro';
import { version } from '../version.mjs';
import { getSections, SITE, SUMMARY, TEXT, toMarkdown } from '../llms';

export const GET: APIRoute = async () => {
	const sections = await getSections();
	const pages = sections.flatMap((section) => section.docs.map(toMarkdown));
	const intro = [
		'# supawho documentation',
		`> ${SUMMARY}`,
		`Complete documentation of supawho ${version} in one file. Website: ${SITE}`,
	].join('\n\n');
	return new Response([intro, ...pages].join('\n\n---\n\n') + '\n', { headers: TEXT });
};
