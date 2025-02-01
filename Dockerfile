FROM golang:1.23.1 AS builder
RUN go install github.com/a-h/templ/cmd/templ@v0.3.833
WORKDIR /go/src/soko
COPY . /go/src/soko
RUN templ generate && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o bin .

FROM node:21 AS assetsbuilder
WORKDIR /go/src/soko
COPY . /go/src/soko
RUN npm install && npx webpack

FROM scratch
WORKDIR /go/src/soko
COPY --from=assetsbuilder /go/src/soko/assets /go/src/soko/assets
COPY --from=builder /go/src/soko/bin /go/src/soko/bin
COPY --from=builder /go/src/soko/pkg /go/src/soko/pkg
COPY --from=builder /go/src/soko/web /go/src/soko/web
ENTRYPOINT ["/go/src/soko/bin/soko", "--serve"]
