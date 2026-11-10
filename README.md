# sso-auto-login

Small program. It waits for an AWS device-login URL, then finishes login in headless Camoufox. The browser starts for that login and stops when it is done.

It is not a web server. It is not part of any other project.

## Build

Needs Go 1.26.1.

```
make build
```

## Config

Copy `.env.example` to `.env` and fill in:

- `CAMOUFOX` — path to the Camoufox binary
- `SSH_HOST` — hostname
- `SSH_USER` — SSH login name (`-l`)
- `SSH_KEY` — path to the private key (`-i`)
- `OKTA_USERNAME` — Okta email
- `OKTA_PASSWORD` — Okta password
- `OKTA_TOTP_SECRET` — the authenticator seed (not a one-time code)

Do not put `.env` in git. `make run` loads it for you.

Cookies go in `camoufox-profile/` in this folder, not your daily browser.

## Run

```
make run
```

It listens on `127.0.0.1:18201` (IPv4 only), then starts:

```
ssh -N -T -l USER -i KEY -o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -R 18201:127.0.0.1:18201 HOST
```

If the VPN is down or the tunnel drops, this program retries SSH (1s, then 2s, up to 30s). Ctrl+C stops Camoufox and SSH, then exits 0.

## User service

```
make install
make restart
make uninstall
```

The user service runs `./sso-auto-login` and loads `.env` from this folder. Uninstall removes the service only.

## Test one URL without SSH

```
set -a && . ./.env && set +a
./sso-auto-login -test-url 'https://revelio.awsapps.com/start/#/device?user_code=TEST-TEST'
```

Or, while it is running:

```
echo 'https://revelio.awsapps.com/start/#/device?user_code=TEST-TEST' | nc -N 127.0.0.1 18201
```

Only two `https` device-login URLs are opened. Other URLs are ignored.

- `https://*.awsapps.com/start/#/device?user_code=`
- `https://<region>.signin.aws/platform/authorize?request_uri=urn:ietf:params:oauth:request_uri:…&identitystore_id=d-…`

## First login

The first Okta page may need selector tweaks if Okta HTML changed. Later logins are faster because cookies stay in the private Camoufox profile.

## Leftover SSH

If you kill this program with `kill -9`, SSH may stay up:

```
make kill-ssh
```
