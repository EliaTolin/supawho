// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// Served at the root of its own domain. SITE_URL is baked in at build time
// (see Dockerfile) and feeds canonical URLs, the sitemap and the JSON-LD.
// `||`, not `??`: the Dockerfile's ARG leaves SITE_URL as an empty string when
// no build argument is passed, and an empty `site` fails the config check.
const site = process.env.SITE_URL || 'https://supawho.auroradigital.it';

export default defineConfig({
	site,
	integrations: [
		starlight({
			title: 'supawho',
			description:
				'Switch between multiple Supabase accounts in seconds — on any OS. Tokens stay in your OS secret vault, never on disk.',
			logo: { src: './src/assets/logo.png', alt: 'supawho' },
			// Self-hosted Umami (cookieless). data-domains keeps local/preview builds
			// from being counted.
			head: [
				{
					tag: 'script',
					attrs: {
						defer: true,
						src: 'https://stats.auroradigital.it/s.js',
						'data-website-id': 'cd8bd58b-1ad9-499d-a02f-8bb3ecbdcf83',
						'data-domains': 'supawho.auroradigital.it',
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
						{ label: 'How it works', slug: 'guides/how-it-works' },
						{ label: 'Security', slug: 'guides/security' },
						{ label: 'Updating', slug: 'guides/updating' },
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
