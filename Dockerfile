# stage 1
# __release_tag__ golang 1.22 was released 2024-02-06
FROM golang:1.22 as build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build

# stage 2: scratch
# __release_tag__ alpine 3.19.1 was released 2024-01-26
FROM alpine:3.19.1 as dist
COPY --from=build /app/bin/pgscv /bin/pgscv
#COPY docker_entrypoint.sh /bin/
EXPOSE 9890
ENTRYPOINT ["/bin/pgscv"]
#ENTRYPOINT ["/bin/docker_entrypoint.sh"]
