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

### Currently...

- [x] Connect DB
- [] Backend
  - [] Logs
    - Obscure sensitive details from logs
  - [] Auth
    - [] Simplify validation logic
    - [] Register
      - [x] Insert user in DB
      - [] Check if user password in https://haveibeenpwned.com
      - [] Add password rules (alphanum, upper+lower, numbers, symbols)
      - [] Send verification email
      - [] Verify Account
    - [] Login
      - [] Check if user in DB
      - [] Compare password
      - [] Token (JWT / Tokens / Cookies)
