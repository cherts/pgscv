# cherts/pgscv-test-runner
# __release_tag__ postrges 14.10 was released 2023-11-09
# __release_tag__ golang 1.22 was released 2024-02-06
# __release_tag__ golangci-lint v1.56.0 was released 2024-02-07
# __release_tag__ gosec v2.18.2 was released 2023-10-23
FROM postgres:14.10

LABEL version="1.0.0"

# install dependencies
RUN apt-get update && \
    apt-get install -y vim make gcc git curl pgbouncer && \
    curl -s -L https://go.dev/dl/go1.22.0.linux-amd64.tar.gz -o - | tar xzf - -C /usr/local && \
    export PATH=$PATH:/usr/local/go/bin && \
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOROOT)/bin v1.56.0 && \
    curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOROOT)/bin v2.18.2 && \
    mkdir /opt/testing/ && \
    rm -rf /var/lib/apt/lists/*

ENV PATH="${PATH}:/usr/local/bin:/usr/local/go/bin"

# copy prepare test environment scripts
COPY prepare-test-environment.sh /usr/bin/
COPY fixtures.sql /opt/testing/

CMD ["echo", "I'm pgscv test runner 1.0.0"]
