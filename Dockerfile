#build stage
FROM golang:alpine AS builder-go
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN apk add --no-cache git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go build -ldflags "-X main.buildVersion=${VERSION} -X main.buildCommit=${COMMIT} -X main.buildDate=${BUILD_DATE}" -o /go/bin/app -v .

#web build stage
FROM node:lts-alpine AS builder-web
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
ENV VITE_APP_VERSION=${VERSION}
ENV VITE_APP_COMMIT=${COMMIT}
ENV VITE_APP_BUILD_DATE=${BUILD_DATE}
WORKDIR /usr/src/app
COPY ./web .
RUN npm ci
RUN npm run build

#final stage
FROM alpine:latest
ARG VERSION=dev
RUN apk --no-cache add ca-certificates
COPY --from=builder-go /go/bin/app /app
COPY --from=builder-web /usr/src/app/build /pb_public

CMD ["/app", "serve", "--http=0.0.0.0:3000"]
LABEL Name=dkpauction Version=${VERSION}
EXPOSE 3000
