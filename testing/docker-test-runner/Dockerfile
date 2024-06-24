# cherts/pgscv-test-runner
# __release_tag__ postrges 14.12 was released 2024-05-09
# __release_tag__ golang 1.22.4 was released 2024-06-04
# __release_tag__ revive v1.3.7 was released 2024-02-07
# __release_tag__ gosec v2.20.0 was released 2024-05-14
FROM postgres:14.12

LABEL version="1.0.1"

# install dependencies
RUN apt-get update && \
    apt-get install -y vim make gcc git curl pgbouncer && \
    curl -s -L https://go.dev/dl/go1.22.4.linux-amd64.tar.gz -o - | tar xzf - -C /usr/local && \
    export PATH=$PATH:/usr/local/go/bin && \
    curl -s -L https://github.com/mgechev/revive/releases/download/v1.3.7/revive_linux_amd64.tar.gz | tar xzf - -C $(go env GOROOT)/bin revive && \
    curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b $(go env GOROOT)/bin v2.20.0 && \
    mkdir /opt/testing/ && \
    rm -rf /var/lib/apt/lists/*

ENV PATH="${PATH}:/usr/local/bin:/usr/local/go/bin"

# copy prepare test environment scripts
COPY prepare-test-environment.sh /usr/bin/
COPY fixtures.sql /opt/testing/

CMD ["echo", "I'm pgscv test runner 1.0.1"]
