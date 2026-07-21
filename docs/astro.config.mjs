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
			social: [
				{ icon: 'github', label: 'GitHub', href: 'https://github.com/EliaTolin/supawho' },
			],
			customCss: ['./src/styles/custom.css'],
			sidebar: [
				{ label: 'Getting started', slug: 'getting-started' },
				{ label: 'Commands', slug: 'commands' },
				{ label: 'Security', slug: 'security' },
			],
		}),
	],
});
