# Stage 1: Build the static site
FROM node:lts-alpine AS build

# Set the working directory
WORKDIR /usr/src/app

# Copy package.json and package-lock.json
COPY package.json package-lock.json ./ 

# Install dependencies
RUN npm ci

# Copy the rest of the application code
COPY . .

# Build the application (produces static files in build/)
RUN npm run build

# Stage 2: Serve the static files with nginx
FROM nginx:alpine

# Copy the built static files to nginx html directory
COPY --from=build /usr/src/app/build /usr/share/nginx/html

# Copy custom nginx config if needed (optional)
# COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]