# ==========================================
# Stage 1: Build VitePress static files
# ==========================================
FROM node:20-alpine AS builder

# Install git (required by VitePress for lastUpdated timestamp feature)
RUN apk add --no-cache git

WORKDIR /app

# Copy dependency manifests
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci

# Copy application source
COPY . .

# Base path argument for deployment flexibility (defaults to root '/' for VPS)
ARG VITEPRESS_BASE=/
ENV VITEPRESS_BASE=${VITEPRESS_BASE}

# Build the static site from the website-only Docker context. The full
# docs:verify gate runs before deployment and additionally requires Go + Chrome.
RUN npm run docs:build-site

# ==========================================
# Stage 2: Serve with Lightweight Nginx
# ==========================================
FROM nginx:1.27-alpine AS runner

# Copy custom Nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built site from builder stage
COPY --from=builder /app/docs/.vitepress/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
