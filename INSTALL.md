## Installing Go

curl -OL https://golang.org/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
rm go1.21.5.linux-amd64.tar.gz

ln -s /usr/local/go/bin/go /usr/bin/go
go version


### Run

go run app.go
