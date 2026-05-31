# =============================================================================
# Dockerfile - MyKanban Frontend (Laxmi)
#
# Multi-stage build: install dependencies & build the Next.js production
# bundle in a full Node image, then serve it with a slim runtime image.
#
# Service name (Compose):  laxmi
# Exposed port:            3000
# =============================================================================

# ---------------------------------------------------------------------------
# Stage 1 — Dependencies
# ---------------------------------------------------------------------------
FROM node:20-alpine AS deps

WORKDIR /app

# Copy only the dependency manifests so Docker can cache the npm install layer
COPY package.json package-lock.json ./

RUN npm ci --frozen-lockfile --only=production

# ---------------------------------------------------------------------------
# Stage 2 — Builder
# ---------------------------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Copy full dependency tree (dev deps included for the build step)
COPY package.json package-lock.json ./
COPY --from=deps /app/node_modules ./node_modules

# Copy source code
COPY . .

# Build the production bundle
RUN npm run build

# ---------------------------------------------------------------------------
# Stage 3 — Runtime
# ---------------------------------------------------------------------------
FROM node:20-alpine AS runner

# Use a non-root user for security
RUN addgroup --system --gid 1001 nodejs \
    && adduser --system --uid 1001 nextjs

WORKDIR /app

# Copy the production build artefacts from the builder
COPY --from=builder /app/public     ./public
COPY --from=builder /app/.next      ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./
COPY --from=builder /app/next.config.mjs ./

# Expose Next.js default port
EXPOSE 3000

ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0

USER nextjs

ENTRYPOINT ["node_modules/.bin/next", "start"]