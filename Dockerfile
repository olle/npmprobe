# Runtime stage
FROM alpine:latest

# Install mlocate (provides locate) and create/update the locate database
RUN apk add --no-cache mlocate && \
    updatedb

# Copy binary to the container
COPY ./bin/npmprobe-linux-x86_64 /usr/local/bin/npmprobe

# Create a sample compromised.txt for testing
RUN mkdir -p /app
WORKDIR /app

# Copy sample compromised packages file
COPY compromised.txt .

# Copy sample package files for testing
RUN mkdir -p ./packages
COPY sample_packages/* ./packages/

RUN updatedb

# Default command: run npmprobe
ENTRYPOINT ["/usr/local/bin/npmprobe"]
CMD ["compromised.txt"]
