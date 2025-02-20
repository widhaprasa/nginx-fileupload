# Step 1: Build the Go app
FROM golang:1.23.6-alpine3.21 as builder

WORKDIR /app
COPY . .

# Build the Go app
RUN go build -o app .

# Step 2: Create the image with Nginx and the Go binary
FROM fabiocicerchia/nginx-lua:1.27.3-alpine3.21.2

# Install supervisor to run multiple processes
RUN apk add --update supervisor && rm  -rf /tmp/* /var/cache/apk/*

# Copy the supervisor configuration
ADD supervisord.conf /etc/

# Copy the compiled Go app from the builder image
COPY --from=builder /app/app /usr/local/bin/

# Expose Nginx default port (80)
EXPOSE 80

# Expose App service port (8080)
EXPOSE 8080

# Command to run using supervisor
ENTRYPOINT ["supervisord", "-n", "-c", "/etc/supervisord.conf"]
