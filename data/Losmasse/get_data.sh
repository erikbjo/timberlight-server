#!/bin/bash

curl -o superficialdeposits_shape.zip https://swift.skyhigh.iik.ntnu.no/swift/v1/c93026420d2d49c69ac937edda870119/superficial_data_public/superficialdeposits_shape.zip

unzip superficialdeposits_shape.zip -d .

rm superficialdeposits_shape.zip
