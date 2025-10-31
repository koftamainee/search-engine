FROM node:22-alpine AS builder

WORKDIR /app

RUN corepack enable && corepack prepare pnpm@latest --activate
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm run build


FROM nginx:stable-alpine

RUN addgroup -g 1000 -S nginxuser && \
  adduser -u 1000 -S nginxuser -G nginxuser

RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /app/dist /usr/share/nginx/html

RUN chown -R nginxuser:nginxuser /var/cache/nginx && \
  chown -R nginxuser:nginxuser /var/log/nginx && \
  chown -R nginxuser:nginxuser /etc/nginx/conf.d

RUN touch /var/run/nginx.pid && \
  chown -R nginxuser:nginxuser /var/run/nginx.pid

USER nginxuser

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
