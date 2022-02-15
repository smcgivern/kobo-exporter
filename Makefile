kobo-exporter: *.go
	@go build

build: kobo-exporter

test: *.go
	go test .

preview: build
	@./kobo_exporter -frequency 30 -config kobo.conf.example

-include *.mk
