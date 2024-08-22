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
    - [] Improve logging with tracing
    - [] Obscure sensitive details from logs
  - [] Auth
    - [x] Simplify validation logic
    - [x] Register
      - [x] Insert user in DB
      - [x] Check if user password in https://haveibeenpwned.com
      - [x] Add password rules (alphanum, upper+lower, numbers, symbols)
      - [] Wrap serial database actions in tx for all-or-none commits
      - [x] Remove `email_verified` column from `users` table
      - [x] Send verification email
      - [x] Verify Account
    - [x] Login
      - [x] Check if user in DB
      - [x] Compare password
      - [x] Token (Paseto ~~JWT~~ / Tokens / Cookies)
    - [] Reset Password
  - [] Tests
    - [] Signup
    - [] Verify Email
    - [] Login
    - [] Reset Password
