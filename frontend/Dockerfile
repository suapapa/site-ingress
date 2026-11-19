# NOTE: This Dockerfile must be built with the project root as context.
# e.g. docker build -f frontend/Dockerfile .

# Build stage
FROM node:20-alpine AS builder
WORKDIR /app

# Copy dependency definitions
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source code
COPY frontend/ .

# Dev uses public/model/go_gopher_low_0.02.bin → ../../asset/...; image needs a real file.
RUN rm -rf public/model
RUN mkdir -p public/model
COPY asset/go_gopher_low_0.02.bin public/model/

# Build the app
RUN npm run build

# Production stage
# Pin the base image instead of floating `nginx:alpine`.
# nginx2root (2026-11-18): CVE-2026-123007 / 123633 / 121810 / 123634 / 113425 affect
# every release up to 1.31.8. Fixed in 1.31.9 (mainline) and 1.30.7 (stable).
FROM nginx:1.31.9-alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
