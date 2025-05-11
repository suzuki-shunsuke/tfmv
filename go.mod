module github.com/suzuki-shunsuke/tfmv

go 1.23.7

toolchain go1.24.3

require (
	github.com/google/go-jsonnet v0.21.0
	github.com/hashicorp/hcl/v2 v2.23.1-0.20250211201033-5c140ce1cb20
	github.com/lintnet/go-jsonnet-native-functions v0.4.2
	github.com/mattn/go-colorable v0.1.14
	github.com/minamijoyo/hcledit v0.2.17
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/afero v1.14.0
	github.com/spf13/pflag v1.0.6
	github.com/suzuki-shunsuke/logrus-error v0.1.4
)

replace github.com/google/go-jsonnet v0.20.0 => github.com/lintnet/go-jsonnet v0.20.2

require (
	github.com/agext/levenshtein v1.2.1 // indirect
	github.com/apparentlymart/go-textseg/v13 v13.0.0 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7 // indirect
	github.com/rogpeppe/go-internal v1.13.1 // indirect
	github.com/zclconf/go-cty v1.13.0 // indirect
	golang.org/x/crypto v0.36.0 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/sys v0.31.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
