# stage 1
# __release_tag__ golang 1.24 was released 2025-02-11
FROM golang:1.24 AS build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build

# stage 2: scratch
# __release_tag__ alpine 3.21.1 was released 2025-02-13
FROM alpine:3.21.3 AS dist
COPY --from=build /app/bin/pgscv /bin/pgscv
#COPY docker_entrypoint.sh /bin/
EXPOSE 9890
ENTRYPOINT ["/bin/pgscv"]
#ENTRYPOINT ["/bin/docker_entrypoint.sh"]
