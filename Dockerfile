FROM golang:1.18 AS build
WORKDIR /pxier_db_syncer
COPY . .
RUN go mod tidy &&  \
    go mod vendor && \
    go build -o pxier_db_syncer && \
    cp config.example.yaml config.yaml

FROM ubuntu:22.04 AS run
COPY --from=build /pxier_db_syncer/pxier_db_syncer .
COPY --from=build /pxier_db_syncer/config.yaml .
CMD ["./pxier_db_syncer"]