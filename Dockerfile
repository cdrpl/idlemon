FROM golang:1.16-alpine AS build
WORKDIR /opt/idlemon
COPY . .
RUN go install .

FROM alpine
EXPOSE 3000
COPY --from=build /go/bin/idlemon /go/bin/idlemon
ENTRYPOINT [ "/go/bin/idlemon", "-e", "nil" ]
