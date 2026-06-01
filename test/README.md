# Local Test Environment

This directory keeps local test-only runtime files for the reader backend.

PostgreSQL:

```bash
cd /Users/cyilin/dev/book-server/test/postgres
sudo docker compose up -d
```

Full test stack:

```bash
cd /Users/cyilin/dev/book-server/test
sudo docker compose up -d --build
```

Backend:

```bash
cd /Users/cyilin/dev/book-server
bash test/run-backend.sh
```

Smoke test, from a second terminal after the backend starts:

```bash
cd /Users/cyilin/dev/book-server
bash test/smoke.sh
```

The backend test data directories are created under:

```text
test/backend-data
```
