# ASK Checker

Checking Danish ASK tax relations for Global ETF's

## Endpoints

- /search - Query for checking ISIN ETFs on the Danish ASK list (`?isni=...`)
- /livez - Liveness endpoint
- /readyz - Readiness endpoint (checks database connection)

## Environment

Database mode is now environment-aware:

- Production (`APP_ENV=production`): defaults to PostgreSQL
- Local testing (`APP_ENV=local`): defaults to SQLite

Optional override:

- Set `DATABASE_TYPE=postgres` or `DATABASE_TYPE=sqlite` to force a specific database.

Required only when using Postgres:

- `DATABASE_URL` must be set when `DATABASE_TYPE=postgres` (or when `APP_ENV=production` and no override is provided).

Other vars:

- `APP_PORT` defaults to `3000`
- `LOG_LEVEL` defaults to `info`


## Build locally

```
go build .
```

```
./ask-checker.exe
```

## Sources

- Danish list of stockbased investment companies - https://skat.dk/erhverv/ekapital/vaerdipapirer/beviser-og-aktier-i-investeringsforeninger-og-selskaber-ifpa
