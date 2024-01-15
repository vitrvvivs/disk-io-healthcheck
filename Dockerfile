FROM golang:1.21 as builder

ENV GOOS linux
ENV CGO_ENABLED 0

WORKDIR /usr/src/app

COPY . .

RUN go build -o disk-io-healthcheck

FROM alpine:3

WORKDIR /usr/src/app 

COPY --from=builder /usr/src/app/disk-io-healthcheck .

EXPOSE 8013

ENTRYPOINT [ "./disk-io-healthcheck" ]
CMD [ "-port", "8013", "-interval", "5", "-device", "/dev/disk/azure/scsi1/lun10", "-max-read", "50000", "-max-write", "50000" ]
