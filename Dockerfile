FROM alpine:3.8

# copy binary into image
COPY ./build/app /app

ENTRYPOINT [ "/app" ]
