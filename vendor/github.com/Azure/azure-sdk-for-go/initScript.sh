set -x
set -e
export GOPATH="$(realpath ../../../..)"
if [ ! -d "$GOPATH/bin" ]
then
  mkdir $GOPATH/bin
fi
curl -sSL https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
PATH=$PATH:$GOPATH/bin
dep ensure
export GO111MODULE=on
go get github.com/Azure/azure-sdk-for-go/eng/tools/generator@latest
cat > $2 << EOF
{
  "envs": {
    "PATH": "$PATH:$GOPATH",
    "GOPATH": "$GOPATH"
  }
}
EOF
