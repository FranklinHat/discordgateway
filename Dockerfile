
# launch go binary

FROM golang:1.22-alpine AS builder

WORKDIR /app

# copy dependency definitions first

COPY go.mod ./

RUN go mod download


#Copy source code, compile statistically

COPY . .
RUN CGO_ENABLED = 0 GOOS=linux go build -o gateway

# Lightweight deployment container

FROM alpine:latest

#Install CA certs for go to call https to discord

RUN apk --no-cache add ca-certificates


WORKDIR /root/

# Copy binary from builder stage

COPY --from=builder /app/gateway .

#Expose port


EXPOSE 8089

#Run binary

CMD ["./gateway"]