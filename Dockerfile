FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

COPY meshdns-linux /meshdns
COPY web/ /web/

EXPOSE 8080

ENTRYPOINT ["/meshdns"]