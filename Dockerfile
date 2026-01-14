#build stage
FROM golang:alpine AS builder-go
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -o /go/bin/app -v .

#web build stage
FROM node:lts-alpine AS builder-web
WORKDIR /usr/src/app
COPY package.json package-lock.json ./ 
RUN npm ci
COPY ./web .
RUN npm run build

#final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder-go /go/bin/app /app
COPY --from=builder-web /usr/src/app/build /pb_public

CMD ["/app", "serve", "--http=0.0.0.0:3000"]
LABEL Name=dkpauction Version=0.0.1
EXPOSE 3000
