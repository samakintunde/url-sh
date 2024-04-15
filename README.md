# URL Shortener

Web: https://url-sh.fly.dev/
API: https://url-sh.fly.dev/api
Health: https://url-sh.fly.dev/health

## Install dependencies

Installs dependencies of the web client

```sh
make install
```

## Start development environment

```sh
make dev
```

## Build app

```sh
  make build
```

### Step 1

- [x] Github repo setup (monorepo)
- [x] Hello world for Go and Solid apps
- [x] Deploy to Fly.io
- [x] Github actions to CI/CD

### Step 2

- [] Connect DB
- [] Backend first
  - [] URL shortener
    - [] Shorten
    - [] Unshorten
  - [] Auth
    - [] Register
      - [] Insert user to DB
      - [] Hash password
      - [] (Email…?)
    - [] Login
      - [] Check if user in DB
      - [] Compare password
      - [] Token (JWT / Tokens / Cookies)
