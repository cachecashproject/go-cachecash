FROM alpine:3.9
RUN apk add caddy
ENTRYPOINT ["caddy"]
CMD ["-port", "80", "-root", "/var/www"]
