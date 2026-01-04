FROM node:22-alpine AS builder

WORKDIR /build

COPY package*.json ./
RUN npm ci --ignore-scripts

COPY . .

ENV VITE_API_URL=REPLACE_ENV_VITE_API_URL
ENV VITE_GAME_ID=REPLACE_ENV_VITE_GAME_ID
ENV VITE_TEST_WIN_ANIMATIONS=REPLACE_ENV_VITE_TEST_WIN_ANIMATIONS

RUN npm run build

FROM nginx:1.27-alpine

COPY --from=builder /build/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY docker-entrypoint.sh /docker-entrypoint.sh

RUN rm -rf /usr/share/nginx/html/*.map && \
    find /usr/share/nginx/html -type f -name "*.js" -exec gzip -k {} \; && \
    find /usr/share/nginx/html -type f -name "*.css" -exec gzip -k {} \; && \
    chmod +x /docker-entrypoint.sh

EXPOSE 80

ENTRYPOINT ["/docker-entrypoint.sh"]

