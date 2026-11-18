import type { APIRoute } from 'astro';
import { version } from '../version.mjs';
import { getSections, markdownUrl, REPO, SITE, SUMMARY, TEXT } from '../llms';

export const GET: APIRoute = async () => {
	const sections = await getSections();
	const lines = [
		'# supawho',
		'',
		`> ${SUMMARY}`,
		'',
		`Latest version: ${version}. Requires the Supabase CLI. Website: ${SITE}`,
		'',
		'Install:',
		'',
		'- macOS: `brew install EliaTolin/tap/supawho`',
		`- Linux: \`curl -fsSL https://raw.githubusercontent.com/EliaTolin/supawho/main/install.sh | bash\``,
		`- Windows (PowerShell): \`irm https://raw.githubusercontent.com/EliaTolin/supawho/main/install.ps1 | iex\``,
		'',
		'Core commands:',
		'',
		'- `supawho add`: save an account (asks for a name and an access token, input hidden)',
		'- `supawho use <name>`: log the Supabase CLI in as that account',
		'- `supawho whoami`: show the email and organizations behind each saved account',
		'- `supawho find <project-ref>`: find which saved account owns a project',
		'',
		`Every page below is plain markdown. The whole documentation in one file: ${SITE}/llms-full.txt`,
	];
	for (const section of sections) {
		lines.push('', `## ${section.title}`, '');
		for (const doc of section.docs) {
			const { title, description } = doc.data;
			lines.push(`- [${title}](${markdownUrl(doc.id)})${description ? `: ${description}` : ''}`);
		}
	}
	lines.push(
		'',
		'## Optional',
		'',
		`- [Source code](${REPO}): Go source, issues and discussions`,
		`- [Releases](${REPO}/releases): changelog and binaries for every version`,
		''
	);
	return new Response(lines.join('\n'), { headers: TEXT });
};
