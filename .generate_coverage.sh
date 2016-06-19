#!/bin/bash -e
# Require installation of: `github.com/wadey/gocovmerge`

cd $GOPATH/src/github.com/kihamo/go-workers

rm -rf ./cov
mkdir cov

i=0
for dir in $(find . -maxdepth 10 -not -path './.git*' -not -path '*/_test.go' -type d);
do
if ls $dir/*.go &> /dev/null; then
    go test -v -coverprofile=./cov/$i.out ./$dir
    i=$((i+1))
fi
done

gocovmerge ./cov/*.out > cover.out
rm -rf ./cov