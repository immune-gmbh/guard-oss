module k6/x/immune

go 1.19

replace (
	github.com/foxboron/go-uefi => github.com/flanfly/go-uefi v0.0.0-20220322115213-c109963105c6
	github.com/google/go-licenses => github.com/immune-gmbh/go-licenses v0.0.0-20220927112129-dc705f009100
	github.com/google/go-tpm => github.com/immune-gmbh/go-tpm v0.3.4-0.20220310140359-93b752e22d71
	// https://github.com/saferwall/pe/pull/47
	// https://github.com/saferwall/pe/pull/48
	github.com/saferwall/pe => github.com/flanfly/pe v0.0.0-20220913101908-93a48b16bf74
)

require (
	github.com/dop251/goja v0.0.0-20230216180835-5937a312edda
	github.com/google/go-tpm v0.3.3
	github.com/google/jsonapi v1.0.0
	github.com/gowebpki/jcs v1.0.0
	github.com/immune-gmbh/agent/v3 v3.12.2
	github.com/sirupsen/logrus v1.9.0
	go.k6.io/k6 v0.43.0
)

require (
	github.com/dlclark/regexp2 v1.8.1 // indirect
	github.com/fatih/color v1.14.1 // indirect
	github.com/go-sourcemap/sourcemap v2.1.4-0.20211119122758-180fcef48034+incompatible // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mstoykov/atlas v0.0.0-20220811071828-388f114305dd // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/onsi/gomega v1.27.1 // indirect
	github.com/oxtoacart/bpool v0.0.0-20190530202638-03653db5a59c // indirect
	github.com/rs/zerolog v1.29.0 // indirect
	github.com/serenize/snaker v0.0.0-20201027110005-a7ad2135616e // indirect
	github.com/spf13/afero v1.9.3 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/guregu/null.v3 v3.5.0 // indirect
)
