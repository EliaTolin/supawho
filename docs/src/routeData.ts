// Per-page <head> additions: social type, JSON-LD and 404 handling.
// Runs for every Starlight page, so structured data covers the whole site
// instead of only the landing page.
import { defineRouteMiddleware } from '@astrojs/starlight/route-data';
import { version } from './version.mjs';

const REPO = 'https://github.com/EliaTolin/supawho';
const MIT = 'https://opensource.org/licenses/MIT';

export const onRequest = defineRouteMiddleware((context) => {
	const route = context.locals.starlightRoute;
	const site = context.site;
	if (!site) return;

	const abs = (path: string) => new URL(path, site).href;
	const ids = {
		website: abs('/#website'),
		author: abs('/#author'),
		software: abs('/#software'),
		source: abs('/#source'),
	};

	// The 404 page must not be indexed and has no canonical URL of its own.
	if (route.id === '404') {
		route.head = route.head.filter(
			(entry) =>
				!(entry.tag === 'link' && entry.attrs?.rel === 'canonical') &&
				!(entry.tag === 'meta' && entry.attrs?.property === 'og:url')
		);
		route.head.push({ tag: 'meta', attrs: { name: 'robots', content: 'noindex' } });
		return;
	}

	const isHome = route.id === '' || route.id === 'index';
	const url = abs(context.url.pathname);
	const { title, description } = route.entry.data;
	let data: Record<string, unknown>;

	if (isHome) {
		// Starlight marks every page as og:type=article; the landing page is a website.
		for (const entry of route.head) {
			if (entry.tag === 'meta' && entry.attrs?.property === 'og:type') {
				entry.attrs.content = 'website';
			}
		}
		data = {
			'@context': 'https://schema.org',
			'@graph': [
				{
					'@type': 'WebSite',
					'@id': ids.website,
					url: abs('/'),
					name: 'supawho',
					description:
						'Documentation for supawho, an open-source CLI to switch between multiple Supabase accounts.',
					inLanguage: 'en',
					publisher: { '@id': ids.author },
				},
				{
					'@type': 'Person',
					'@id': ids.author,
					name: 'Elia Tolin',
					url: 'https://github.com/EliaTolin',
					sameAs: ['https://github.com/EliaTolin'],
				},
				{
					'@type': 'SoftwareApplication',
					'@id': ids.software,
					name: 'supawho',
					url: abs('/'),
					description:
						'Free, open-source CLI to save multiple Supabase accounts and switch between them. Access tokens are stored in the operating system secret vault.',
					applicationCategory: 'DeveloperApplication',
					operatingSystem: 'macOS, Linux, Windows',
					softwareVersion: version,
					image: abs('/og.png'),
					downloadUrl: `${REPO}/releases`,
					installUrl: abs('/guides/getting-started/'),
					license: MIT,
					isAccessibleForFree: true,
					author: { '@id': ids.author },
					publisher: { '@id': ids.author },
					sameAs: [REPO],
					offers: { '@type': 'Offer', price: '0', priceCurrency: 'USD' },
				},
				{
					'@type': 'SoftwareSourceCode',
					'@id': ids.source,
					name: 'supawho source code',
					codeRepository: REPO,
					programmingLanguage: 'Go',
					license: MIT,
					author: { '@id': ids.author },
					targetProduct: { '@id': ids.software },
				},
			],
		};
	} else {
		const crumbs = [{ name: 'supawho', item: abs('/') }];
		if (route.id.startsWith('commands/')) {
			crumbs.push({ name: 'Commands', item: abs('/commands/') });
		}
		crumbs.push({ name: title, item: url });

		data = {
			'@context': 'https://schema.org',
			'@graph': [
				{
					'@type': 'TechArticle',
					'@id': `${url}#article`,
					headline: title,
					...(description ? { description } : {}),
					url,
					mainEntityOfPage: url,
					inLanguage: 'en',
					image: abs('/og.png'),
					author: { '@type': 'Person', '@id': ids.author, name: 'Elia Tolin', url: 'https://github.com/EliaTolin' },
					publisher: { '@id': ids.author },
					isPartOf: { '@id': ids.website },
					about: { '@id': ids.software },
					...(route.lastUpdated ? { dateModified: route.lastUpdated.toISOString().slice(0, 10) } : {}),
				},
				{
					'@type': 'BreadcrumbList',
					itemListElement: crumbs.map((crumb, i) => ({
						'@type': 'ListItem',
						position: i + 1,
						...crumb,
					})),
				},
			],
		};
	}

	// Point AI agents at the plain-markdown copy of the page (see src/llms.ts).
	if (!isHome) {
		route.head.push({
			tag: 'link',
			attrs: { rel: 'alternate', type: 'text/markdown', href: abs(`/${route.id}.md`) },
		});
	}

	route.head.push({
		tag: 'script',
		attrs: { type: 'application/ld+json' },
		// Escape "<" so page text can never close the script element.
		content: JSON.stringify(data).replace(/</g, '\\u003c'),
	});
});
