all:
	go build -o otc
	go test -count=1

bench:
	go build -o otc
	go test ./... --bench .
