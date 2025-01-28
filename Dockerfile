# stage 1
# __release_tag__ golang 1.23 was released 2024-08-13
FROM golang:1.23 AS build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build

# stage 2: scratch
# __release_tag__ alpine 3.20.0 was released 2024-05-22
FROM alpine:3.20.0 AS dist
COPY --from=build /app/bin/pgscv /bin/pgscv
#COPY docker_entrypoint.sh /bin/
EXPOSE 9890
ENTRYPOINT ["/bin/pgscv"]
#ENTRYPOINT ["/bin/docker_entrypoint.sh"]
