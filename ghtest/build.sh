#!/bin/bash
############################################

# File Name : build.sh

# Purpose :

# Creation Date : 04-06-2016

# Last Modified : Tue May 24 16:42:22 2016

# Created By : Kiyor 

############################################

env GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'main.buildtime=$(date -u '+%Y%m%d%H%M%S')'"
\cp ./ghtest ./ghtest_osx
env GOOS=linux GOARCH=amd64 go build -ldflags "-X 'main.buildtime=$(date -u '+%Y%m%d%H%M%S')'"
\cp ./ghtest ./ghtest_linux
rm -f ./ghtest
if [[ $(uname) -eq "Darwin" ]];then
	ln -s ./ghtest_osx ./ghtest
else
	ln -s ./ghtest_linux ./ghtest
fi
