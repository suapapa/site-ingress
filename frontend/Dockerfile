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

# Dev uses public/model/go_gopher_low_0.02.stl → ../../asset/...; image needs a real file.
RUN rm -rf public/model
RUN mkdir -p public/model
COPY asset/go_gopher_low_0.02.stl public/model/

# Build the app
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY frontend/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
