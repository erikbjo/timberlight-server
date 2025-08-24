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

unzip "$1"/superficialdeposits_shape.zip -d "$1"

mv "$1"/Losmasse/* "$1" 
rm "$1"/Losmasse

rm "$1"/superficialdeposits_shape.zip
