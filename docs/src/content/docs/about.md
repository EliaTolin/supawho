---
title: About supawho
description: Who maintains supawho, how it is licensed, how it relates to Supabase, and where to report bugs or contribute.
lastUpdated: 2026-11-17
---

supawho is a free, open-source command-line tool for switching between multiple Supabase accounts. It is written in Go and released under the [MIT license](https://github.com/EliaTolin/supawho/blob/main/LICENSE).

## Maintainer

supawho is created and maintained by **Elia Tolin** ([@EliaTolin](https://github.com/EliaTolin) on GitHub).

The project became cross-platform thanks to the community: [@igorgomes3](https://github.com/igorgomes3)'s [supawho-linux](https://github.com/igorgomes3/supawho-linux) fork sparked the rewrite that now runs on macOS, Linux and Windows.

## Relationship with Supabase

supawho is an independent community project. It is not affiliated with or endorsed by Supabase. It works by driving the official [Supabase CLI](https://supabase.com/docs/guides/cli) and calling the public Supabase Management API.

## Source, releases and changelog

- Source code: [github.com/EliaTolin/supawho](https://github.com/EliaTolin/supawho)
- Releases and release notes: [GitHub releases](https://github.com/EliaTolin/supawho/releases)
- Full history: [CHANGELOG.md](https://github.com/EliaTolin/supawho/blob/main/CHANGELOG.md)

Every release ships a `checksums.txt` file, which [`supawho upgrade`](/commands/upgrade/) verifies before replacing the binary.

## Contributing and support

Bug reports, ideas and pull requests are welcome in the [issue tracker](https://github.com/EliaTolin/supawho/issues). Before opening an issue, check [Troubleshooting](/guides/troubleshooting/).

## This website

The documentation is built with [Starlight](https://starlight.astro.build) and lives in the `docs/` folder of the same repository. It uses self-hosted, cookieless analytics (Umami) and sets no tracking cookies.
