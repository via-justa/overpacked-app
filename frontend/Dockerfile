# Build stage
# Pin the builder to the build host's platform (BUILDPLATFORM): node:*-alpine has
# no linux/arm/v7 manifest, and the build only emits architecture-independent static
# assets, so we always build natively and copy the result into the target-arch runtime.
FROM --platform=$BUILDPLATFORM node:24-alpine AS builder

WORKDIR /app

# API base URL baked into the bundle. EMPTY for the standard deployment: nginx
# serves the SPA and reverse-proxies /api to the backend on the same origin, and
# the client request paths already include the /api/v1 prefix — so a non-empty
# value here (e.g. "/api") would double it to /api/api/v1 and 404. Override only
# to point the SPA at a different ORIGIN (e.g. https://api.example.com).
ARG VITE_API_BASE_URL=
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}

# Copy package files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm ci --ignore-scripts

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Runtime stage
FROM nginx:1.31-alpine

# Use nginx envsubst template so backend upstream can be overridden at runtime.
COPY nginx.conf /etc/nginx/templates/default.conf.template
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
