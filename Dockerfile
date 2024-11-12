# Step 1: Build the Go app
FROM golang:1.21.13-bookworm as builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go build -o app .

# Step 2: Create the image with Nginx and the Go binary
FROM nginx:1.27.2-bookworm

# Copy the compiled Go app from the builder image
COPY --from=builder /app/app /usr/local/bin/

# Expose Nginx default port (80)
EXPOSE 80

# Expose the Go service port (8080)
EXPOSE 8080

# Command to run both Go app and Nginx
CMD ["sh", "-c", "/usr/local/bin/app & nginx -g 'daemon off;'"]
