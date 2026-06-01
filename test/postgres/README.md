# Test PostgreSQL

This directory contains a local PostgreSQL test instance for the reader backend.

All database files are kept under:

```text
test/postgres/data
```

Start:

```bash
cd test/postgres
docker compose up -d
```

Backend DSN:

```text
postgres://book_reader:book_reader_password@localhost:15432/book_reader?sslmode=disable
```

Stop:

```bash
docker compose down
```
