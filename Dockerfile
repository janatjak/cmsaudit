FROM golang as build

WORKDIR /go/src/github.com/janatjak/cmsaudit
COPY . .
RUN go build


FROM debian:11-slim

RUN apt update && apt install -y ca-certificates
COPY --from=build /go/src/github.com/janatjak/cmsaudit/cmsaudit /usr/local/bin

CMD ["cmsaudit"]
