# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

# Copy package and lock files
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install

# Copy all source files
COPY . .

# Build the application using Vite
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built files to nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Optional: Copy a custom nginx configuration file if required
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose the default nginx port
EXPOSE 80

# Start the nginx server
CMD ["nginx", "-g", "daemon off;"]
