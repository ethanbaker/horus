# syntax=docker/dockerfile:1

FROM golang:1.20
ENV GO111MODULE=auto

# Set the working directory for the docker image
WORKDIR /go/src/app

# Copy the build information over
COPY . .

# Install bot modules
WORKDIR /go/src/app/bot
RUN go mod download

# Install utils modules
WORKDIR /go/src/app/utils
RUN go mod download

# Build the application to an executable
WORKDIR /go/src/app
RUN go build -o discord_main ./implementations/discord

# Run the executable
CMD [ "/go/src/app/discord_main" ]

# Commands to build and run:
# docker build --tag horus .
# docker run -d --network="host" --name horus --env-file config/.env horus
