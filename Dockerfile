﻿# stage 1
# __release_tag__ golang 1.25 was released 2025-08-12
FROM golang:1.25 AS build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build

# stage 2: scratch
# __release_tag__ alpine 3.22.0 was released 2025-05-30
FROM alpine:3.22.0 AS dist
COPY --from=build /app/bin/pgscv /bin/pgscv
#COPY docker_entrypoint.sh /bin/
EXPOSE 9890
#EXPOSE 6060
ENTRYPOINT ["/bin/pgscv"]
#ENTRYPOINT ["/bin/docker_entrypoint.sh"]
