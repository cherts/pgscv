# stage 1
# __release_tag__ golang 1.23 was released 2024-08-13
FROM golang:1.23 AS build
LABEL stage=intermediate
WORKDIR /app
COPY . .
RUN make build
EXPOSE 9890
EXPOSE 8080
ENTRYPOINT ["/bin/pgscv"]
