# ---- build stage ----
# Astro 7 requires Node >= 22.12; pin an image well above that so the build
# never depends on whatever the platform's builder happens to resolve.
FROM node:24-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Baked into canonical URLs and the sitemap at build time.
ARG SITE_URL
ENV SITE_URL=$SITE_URL
RUN npm run build

# ---- serve stage ----
FROM nginx:1.27-alpine

COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
