# s3-storage — S3-compatible API (MinIO locally)

- **Domain port:** `ObjectStorage` in `internal/domain`
- **Infrastructure:** AWS SDK v2 `service/s3` with custom endpoint (MinIO / AWS)

## MinIO

```bash
docker compose up -d
```

Console: http://localhost:9001 (user/pass `minioadmin` / `minioadmin` by default).

## Run API

```bash
go run ./cmd/api
```

## Example

```bash
curl -T README.md "http://localhost:8080/objects/hello.txt"
curl "http://localhost:8080/objects/hello.txt"
```

## Env

See `.env.example` (`S3_ENDPOINT`, `S3_BUCKET`, keys, region).
