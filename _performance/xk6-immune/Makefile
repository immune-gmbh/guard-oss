all: k6

k6: main.go go.mod Makefile
	xk6 build --with k6/x/immune=. --replace github.com/google/go-tpm=github.com/immune-gmbh/go-tpm@v0.3.4-0.20220310140359-93b752e22d71 --replace github.com/immune-gmbh/agent/v3=../../../../agent
