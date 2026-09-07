# FARO audit explorer image.
#
# The explorer is a static site (next.config.ts sets output: 'export'): it
# reads the log over HTTP from the browser and holds no server-side state.
# Containerizing it is only about giving it the same deploy path as the log
# service, not about running a Node server in production — the runtime stage
# is nginx serving the exported files, nothing else.
#
# NEXT_PUBLIC_FARO_URL is baked in at build time: Next.js inlines
# NEXT_PUBLIC_* variables into the static output, so it cannot be changed by
# an environment variable at container start.

FROM node:22-alpine AS build
WORKDIR /src

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
ARG NEXT_PUBLIC_FARO_URL=http://localhost:2025
ENV NEXT_PUBLIC_FARO_URL=${NEXT_PUBLIC_FARO_URL}
RUN npm run build

# Unprivileged image: runs as a non-root user and listens on 8080 by
# default, so the container needs no special capabilities.
FROM nginxinc/nginx-unprivileged:1.27-alpine
COPY --from=build /src/out /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 8080
