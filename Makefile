.PHONY: build run test fmt clean kill-ssh install uninstall restart

BINARY := sso-auto-login
UNITDIR := $(HOME)/.config/systemd/user
UNIT := sso-auto-login.service

build:
	go build -o $(BINARY) ./cmd/sso-auto-login

run: build
	set -a && . ./.env && set +a && ./$(BINARY)

test:
	go test -race -shuffle=on -count=1 ./...

fmt:
	gofmt -w .

clean:
	rm -f $(BINARY)

kill-ssh:
	pkill -f 'ssh -N.*18201' || true

install: build
	mkdir -p $(UNITDIR)
	sed 's|@DIR@|$(CURDIR)|g' $(UNIT) > $(UNITDIR)/$(UNIT)
	systemctl --user daemon-reload
	systemctl --user enable --now $(UNIT)

uninstall:
	systemctl --user disable --now $(UNIT) || true
	rm -f $(UNITDIR)/$(UNIT)
	systemctl --user daemon-reload

restart: build
	systemctl --user restart $(UNIT)
