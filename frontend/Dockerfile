
FROM node:22-alpine AS dependencies
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS builder
WORKDIR /app
COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
ARG API_PORT=8080
# Keep the legacy browser API/WebSocket URLs aligned with Compose's host port.
RUN case "$API_PORT" in ''|*[!0-9]*) exit 1 ;; esac \
    && test "$API_PORT" -ge 1 && test "$API_PORT" -le 65535 \
    && find app components utils -type f -exec sed -i "s/localhost:8080/localhost:${API_PORT}/g" {} +
RUN npm run build && npm prune --omit=dev --ignore-scripts

FROM node:22-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production

COPY --from=builder /app/package.json /app/package-lock.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next ./.next

EXPOSE 3000
CMD ["npm", "run", "start"]
