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

# Handle the symlinked asset:
# The source code checks in public/model as a symlink to ../../asset/go-gopher-model.
# In the container, we need the actual files.
# We remove the symlink and copy the actual assets from the build context.
RUN rm public/model || true
COPY asset/go-gopher-model public/model

# Build the app
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
