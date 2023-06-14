FROM golang:1.20-alpine

# Install necessary dependencies
RUN apk add --no-cache bash

RUN apk add --no-cache make

# Copy your project files
COPY . /traderui

# Set the working directory
WORKDIR /traderui

# Build the application
RUN make build

# Specify the command to run when the container starts
CMD ["./bin/traderui"]
