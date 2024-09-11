# pgSCV demo laboratory

### Requirements

- Docker
- Docker Compose

### List of running containers and ports
- pgscv (listen port: 9890)
- grafana (listen port: 3000)
- vmagent (listen port: 8429)
- victoriametrics (listen port: 8428)
- postgres12 (listen port: 5432)
- postgres13 (listen port: 5433)
- postgres14 (listen port: 5434)
- postgres15 (listen port: 5435)
- postgres16 (listen port: 5436)
- pgbouncer12 (listen port: 6432)
- pgbouncer13 (listen port: 6433)
- pgbouncer14 (listen port: 6434)
- pgbouncer15 (listen port: 6435)
- pgbouncer16 (listen port: 6436)
- etcd1
- etcd2
- etcd3
- patroni1 (listen port: 5437, 8008)
- patroni2 (listen port: 5438, 8009)
- patroni3 (listen port: 5439, 8010)
- haproxy (listen port: 5000, 5001)
- pgbench_12
- pgbench_13
- pgbench_14
- pgbench_15
- pgbench_16
- pgbench_patroni
- pgbench_patroni_s

### Quick start

Run pgSCV demo laboratory:
```bash
docker-compose up -d
```

Start pgbench tests:
```bash
./start_pgbench.sh
```

Open Grafana into Web browser URL: http://127.0.0.1:3000

Login: admin

Password: admin

Open pgSCV dashboards, enjoy and drink coffee ;)

View pgSCV logs:
```bash
docker logs pgscv -f
```

### Stop demo laboratory and cleanup data

Stop pgbench tests:
```bash
./stop_pgbench.sh
```

Stop pgSCV demo laboratory and cleanup demo data:
```bash
./stop_and_cleanup_data.sh
```
