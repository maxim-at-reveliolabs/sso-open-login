# AGENTS.md

Local Go 1.26.1 daemon. One binary. Not part of any other repo.

It listens on `127.0.0.1:18201` (IPv4 TCP, not HTTP), keeps an SSH reverse tunnel up, and opens only AWS IAM Identity Center device-login URLs in Camoufox (headless, started per login).

## Commands

- `make build` — `./sso-auto-login`
- `make run` — load `.env`, run in the foreground
- `make test` — `go test -race -shuffle=on -count=1 ./...`
- `make install` / `make restart` / `make uninstall` — systemd user unit
- `make kill-ssh` — leftover `ssh -N` after `kill -9`

## Layout

- `cmd/sso-auto-login` — `os.Exit` only
- `internal/app` — flags, listen loop, one-line TCP protocol
- `internal/allowurl` — URL allow list (parse with `net/url`, use `Hostname()`)
- `internal/config` — env only, no JSON
- `internal/tunnel` — SSH child + retry while VPN is down
- `internal/browser` — playwright-go + Camoufox, one login at a time, 3 minute timeout

## Config

Read from the process environment (Makefile / user service load `.env`):

`CAMOUFOX`, `SSH_HOST`, `SSH_USER`, `SSH_KEY`, `OKTA_USERNAME`, `OKTA_PASSWORD`, `OKTA_TOTP_SECRET`

Camoufox profile: `camoufox-profile/` next to the binary. Never the daily browser. Binary default: `/home/maxim/.cache/camoufox/camoufox-bin`.

User service runs the binary and `.env` in this folder.

## Rules

- Do not print password, TOTP secret, or TOTP code.
- Do not add HTTP, a system tray, Brave, Tampermonkey, or extra features.
- Use Camoufox via playwright-go, headless. Start it when a URL arrives and close it when login finishes. `CAMOUFOX` is the path.
- Open only `https` `*.awsapps.com` with path `/start` and fragment `/device?user_code=`. Reject userinfo, `http`, `file`, `javascript`.
- SSH: `ssh -N -T -l USER -i KEY -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -R 18201:127.0.0.1:18201 HOST`. USER and KEY come from `SSH_USER` and `SSH_KEY`. No passwords in this program. Restart on drop; do not exit the daemon because SSH died. Ctrl+C still kills SSH and exits 0.
- Okta HTML changes. Prefer waiting for selectors (classic + OIE). Log the failed step name, not field values.
- 250 pure LOC per file. `slog` to stderr. No new deps if stdlib or playwright-go/otp already cover it.
