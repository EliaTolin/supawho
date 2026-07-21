// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// Project site on GitHub Pages (custom domain): https://www.eliatolin.it/supawho
export default defineConfig({
	site: 'https://www.eliatolin.it',
	base: '/supawho',
	integrations: [
		starlight({
			title: 'supawho',
			description:
				'Switch between multiple Supabase accounts in seconds — on any OS. Tokens stay in your OS secret vault, never on disk.',
			logo: { src: './src/assets/logo.png', alt: 'supawho' },
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
