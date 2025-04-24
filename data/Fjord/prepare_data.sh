#!/bin/bash

if [ $# -eq 0 ];
then
  echo "$0: Missing arguments"
  exit 1
elif [ $# -gt 1 ];
then
  echo "$0: Too many arguments: $@"
  exit 1
fi

echo "Using path: $1"

unzip "$1"/fjordkatalogen_omrade.zip -d "$1"

rm "$1"/fjordkatalogen_omrade.zip
