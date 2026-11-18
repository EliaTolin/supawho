// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';
import sitemap from '@astrojs/sitemap';
import { readFileSync, readdirSync } from 'node:fs';
import { join, relative } from 'node:path';

// Served at the root of its own domain. SITE_URL is baked in at build time
// (see Dockerfile) and feeds canonical URLs, the sitemap and the JSON-LD.
// `||`, not `??`: the Dockerfile's ARG leaves SITE_URL as an empty string when
// no build argument is passed, and an empty `site` fails the config check.
const site = process.env.SITE_URL || 'https://supawho.dev';

// Sitemap <lastmod> comes from each page's `lastUpdated` frontmatter. The
// Docker build context is `docs/` with no `.git`, so git dates aren't
// available: bump `lastUpdated` when you edit a page.
const docsDir = new URL('./src/content/docs/', import.meta.url).pathname;
/** @type {Map<string, string>} */
const lastmod = new Map();
/** @param {string} dir */
function collectDates(dir) {
	for (const entry of readdirSync(dir, { withFileTypes: true })) {
		const path = join(dir, entry.name);
		if (entry.isDirectory()) collectDates(path);
		else if (/\.mdx?$/.test(entry.name)) {
			const date = readFileSync(path, 'utf8').match(/^lastUpdated:\s*(\d{4}-\d{2}-\d{2})/m)?.[1];
			if (!date) continue;
			const slug = relative(docsDir, path)
				.replace(/\.mdx?$/, '')
				.replace(/(^|\/)index$/, '');
			lastmod.set(slug ? `/${slug}/` : '/', date);
		}
	}
}
collectDates(docsDir);

const ogImage = new URL('/og.png', site).href;

export default defineConfig({
	site,
	integrations: [
		// Declared before Starlight so it is used instead of Starlight's default
		// sitemap integration.
		sitemap({
			serialize(item) {
				const date = lastmod.get(new URL(item.url).pathname);
				if (date) item.lastmod = date;
				return item;
			},
		}),
		starlight({
			title: 'supawho',
			description:
				'Switch between multiple Supabase accounts in seconds, on any OS. Access tokens are kept in your OS secret vault.',
			// Decorative: the site title next to it already names the project.
			logo: { src: './src/assets/logo.webp', alt: '' },
			favicon: '/favicon.png',
			routeMiddleware: './src/routeData.ts',
			// Self-hosted Umami (cookieless). data-domains keeps local/preview builds
			// from being counted.
			head: [
				{ tag: 'link', attrs: { rel: 'icon', href: '/favicon.ico', sizes: '32x32' } },
				{ tag: 'link', attrs: { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' } },
				{ tag: 'meta', attrs: { property: 'og:image', content: ogImage } },
				{ tag: 'meta', attrs: { property: 'og:image:width', content: '1200' } },
				{ tag: 'meta', attrs: { property: 'og:image:height', content: '630' } },
				{
					tag: 'meta',
					attrs: {
						property: 'og:image:alt',
						content: 'supawho: switch between multiple Supabase accounts',
					},
				},
				{ tag: 'meta', attrs: { name: 'twitter:image', content: ogImage } },
				{
					tag: 'script',
					attrs: {
						defer: true,
						src: 'https://stats.auroradigital.it/s.js',
						'data-website-id': 'cd8bd58b-1ad9-499d-a02f-8bb3ecbdcf83',
						'data-domains': 'supawho.dev',
					},
				},
			],
			components: {
				ThemeSelect: './src/components/ThemeSelect.astro',
				ThemeProvider: './src/components/ThemeProvider.astro',
			},
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/EliaTolin/supawho' },
			],
			customCss: [
				'@fontsource-variable/inter',
				'@fontsource-variable/source-code-pro',
				'./src/styles/custom.css',
			],
			sidebar: [
				{
					label: 'Guides',
					items: [
						{ label: 'Introduction', slug: 'guides/introduction' },
						{ label: 'Getting started', slug: 'guides/getting-started' },
						{ label: 'Multiple Supabase accounts', slug: 'guides/multiple-supabase-accounts' },
						{ label: 'supawho vs. alternatives', slug: 'guides/alternatives' },
						{ label: 'How it works', slug: 'guides/how-it-works' },
						{ label: 'Security', slug: 'guides/security' },
						{ label: 'CI and automation', slug: 'guides/ci-and-automation' },
						{ label: 'Troubleshooting', slug: 'guides/troubleshooting' },
						{ label: 'Updating', slug: 'guides/updating' },
						{ label: 'About', slug: 'about' },
					],
				},
				{
					label: 'Commands',
					items: [
						{ label: 'Overview', slug: 'commands' },
						{ label: 'supawho (interactive)', slug: 'commands/interactive' },
						{ label: 'add', slug: 'commands/add' },
						{ label: 'use', slug: 'commands/use' },
						{ label: 'list', slug: 'commands/list' },
						{ label: 'whoami', slug: 'commands/whoami' },
						{ label: 'find', slug: 'commands/find' },
						{ label: 'rename', slug: 'commands/rename' },
						{ label: 'remove', slug: 'commands/remove' },
						{ label: 'upgrade', slug: 'commands/upgrade' },
						{ label: 'version', slug: 'commands/version' },
					],
				},
			],
		}),
	],
});
