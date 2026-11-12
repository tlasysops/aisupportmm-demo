# Sample App (frontend + backend containers)

- Frontend: https://weblayer.aisupportmm.com  (nginx static page, env `API_URL`)
- Backend:  https://apilayer.aisupportmm.com  (Go API + existing PostgreSQL, env `DATABASE_URL`, `ALLOWED_ORIGIN`)
- CI/CD: `.github/workflows/deploy.yml` (test -> build -> push GHCR -> deploy CapRover)

## Local
    docker compose up --build   # frontend :8080, backend :8081

## Production env vars (CapRover)
Backend app:  PORT=8080, DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=disable,
              ALLOWED_ORIGIN=https://weblayer.aisupportmm.com   | container HTTP port 8080
Frontend app: API_URL=https://apilayer.aisupportmm.com           | container HTTP port 80

## GitHub
Variables: CAPROVER_SERVER, BACKEND_APP, FRONTEND_APP
Secrets:   BACKEND_DEPLOY_TOKEN, FRONTEND_DEPLOY_TOKEN
