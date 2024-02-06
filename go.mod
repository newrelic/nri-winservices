module github.com/newrelic/nri-winservices

go 1.21.6

require (
	github.com/newrelic/infra-integrations-sdk/v4 v4.2.1
	// The exporter version packaged with the integration is defined in win_build.ps1
	github.com/prometheus-community/windows_exporter v0.25.1
	github.com/prometheus/client_model v0.5.0
	github.com/prometheus/common v0.46.0
	github.com/stretchr/testify v1.8.4
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/sys v0.16.0
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/alecthomas/kingpin/v2 v2.4.0 // indirect
	github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-kit/log v0.2.1 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/newrelic/infrastructure-agent v0.0.0-20201127092132-00ac7efc0cc6 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.18.0 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xhit/go-str2duration/v2 v2.1.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
