// Plain-markdown copies of the docs for LLMs and AI agents: /llms.txt (index),
// /llms-full.txt (everything in one file) and /<page>.md (one page each).
// Built from the same content collection as the site, so they never drift.
import { getCollection, type CollectionEntry } from 'astro:content';

export const SITE = 'https://supawho.dev';
export const REPO = 'https://github.com/EliaTolin/supawho';

export const SUMMARY =
	'supawho is a free, open-source CLI (MIT, written in Go) that saves multiple Supabase accounts and switches the Supabase CLI between them with one command, on macOS, Linux and Windows. Access tokens are stored in the operating system secret vault (Keychain, Secret Service, Credential Manager). supawho is a community project, not affiliated with Supabase.';

// Same order as the sidebar.
const SECTIONS: { title: string; ids: string[] }[] = [
	{
		title: 'Guides',
		ids: [
			'guides/introduction',
			'guides/getting-started',
			'guides/multiple-supabase-accounts',
			'guides/alternatives',
			'guides/how-it-works',
			'guides/security',
			'guides/ci-and-automation',
			'guides/troubleshooting',
			'guides/updating',
		],
	},
	{
		title: 'Commands',
		ids: [
			'commands',
			'commands/interactive',
			'commands/add',
			'commands/use',
			'commands/list',
			'commands/whoami',
			'commands/find',
			'commands/projects',
			'commands/doctor',
			'commands/rename',
			'commands/remove',
			'commands/upgrade',
			'commands/version',
		],
	},
	{ title: 'Project', ids: ['about'] },
];

type Doc = CollectionEntry<'docs'>;

export const pageUrl = (id: string) => `${SITE}/${id}/`;
export const markdownUrl = (id: string) => `${SITE}/${id}.md`;

/** Sections with their entries, failing the build if a listed page is missing or unlisted. */
export async function getSections() {
	const docs = await getCollection('docs');
	const byId = new Map(docs.map((doc) => [doc.id, doc]));
	const listed = new Set(SECTIONS.flatMap((section) => section.ids));
	const unlisted = docs.filter(
		(doc) => !listed.has(doc.id) && doc.id !== 'index' && doc.id !== '404'
	);
	if (unlisted.length) {
		throw new Error(`Add to SECTIONS in src/llms.ts: ${unlisted.map((d) => d.id).join(', ')}`);
	}
	return SECTIONS.map((section) => ({
		title: section.title,
		docs: section.ids.map((id) => {
			const doc = byId.get(id);
			if (!doc) throw new Error(`src/llms.ts lists a page that does not exist: ${id}`);
			return doc;
		}),
	}));
}

/** One page as standalone markdown: title, summary, canonical URL, body. */
export function toMarkdown(doc: Doc): string {
	const body = (doc.body ?? '')
		// Site-relative links only make sense on the site.
		.replace(/\]\(\/(?!\/)/g, `](${SITE}/`)
		// Starlight asides -> plain markdown.
		.replace(/^:::(note|tip|caution|danger)(?:\[(.*?)\])?\s*$/gm, (_m, kind: string, label?: string) =>
			`**${label ?? kind[0].toUpperCase() + kind.slice(1)}:**`
		)
		.replace(/^:::\s*$/gm, '')
		// Code-fence options are for the site renderer.
		.replace(/^(```[\w-]+)[ \t]+\S.*$/gm, '$1')
		.replace(/\n{3,}/g, '\n\n')
		.trim();
	const { title, description } = doc.data;
	return [
		`# ${title}`,
		description ? `> ${description}` : '',
		`Source: ${pageUrl(doc.id)}`,
		body,
	]
		.filter(Boolean)
		.join('\n\n');
}

export const TEXT = { 'Content-Type': 'text/plain; charset=utf-8' };
export const MARKDOWN = { 'Content-Type': 'text/markdown; charset=utf-8' };
