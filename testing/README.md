# pgSCV - Testing environment

### Requirements
- Debian-like OS (Debian, Ubuntu)
- Docker

### Quick start

#### Testing pgSCV on PostgreSQL v18

Prepare environment in docker container:
```bash
git clone https://github.com/cherts/pgscv.git
cd pgscv
docker run -it --rm --name pgscv_18 -v ${PWD}:/opt/pgscv -p 9890:9890 postgres:18 bash
cd /opt/pgscv
testing/prepare_os_environment.sh 18
testing/prepare_test_environment.sh 18
export PATH=$PATH:/usr/local/bin:/usr/local/go/bin
```

Run tests:
```bash
make test
```

Build and run local pgSCV:
```bash
make build
./bin/pgscv --config-file="testing/pgscv.yaml" --log-level="debug"
```

Flush environment (stop and delete all postgres instance):
```bash
testing/flush_test_environment.sh 18
```