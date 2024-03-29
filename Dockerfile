FROM golang:alpine as build

WORKDIR /go/src/github.com/janatjak/cmsaudit
COPY . .
RUN CGO_ENABLED=0 go build


FROM alpine
COPY --from=build /go/src/github.com/janatjak/cmsaudit/cmsaudit /usr/local/bin

CMD ["cmsaudit"]
