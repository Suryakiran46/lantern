build:
	go build -o lantern .

install:
	go install .

clean:
	rm -f lantern

cross-linux:
	GOOS=linux GOARCH=amd64 go build -o lantern-linux .

cross-windows:
	GOOS=windows GOARCH=amd64 go build -o lantern.exe .