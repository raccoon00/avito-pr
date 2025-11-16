FROM scratch

WORKDIR /app
COPY ./bin/restapi /app
CMD ["/app/restapi"]
