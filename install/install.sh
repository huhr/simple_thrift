#! /bin/bash

GOROOT=/usr/local/go

if [ ! -f install.sh ]; then
    echo 'install.sh must be run within its container folder' 1>&2
	exit 1
fi

CURRDIR=`pwd`/../
FAILED=0

export GOPATH=$CURRDIR

echo 'Compiling and installing...'
go clean hello
go install hello

if (($? != 0))
then
	FAILED=1
fi

if (( $FAILED == 0 ))
then
	echo 'Done'
else
	echo 'Error'
	exit 1
fi
