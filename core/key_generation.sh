#!/bin/sh

export KEY="/key"
mkdir -p ${KEY}

# Install python
apk update  
apk add python3
apk add py3-pip
apk add musl-dev

apk add libffi-dev
apk add --no-cache gcc
apk add autoconf automake build-base libtool pkgconfig python3-dev

# Install required libraries
echo "Install python libraries..."
pip3 install wheel 

pip3 install bip_utils 
pip3 install coincurve --no-binary coincurve 
pip3 install eciespy 
# Execute the script to generate the keys

echo "Key generation..."
keys=`python3 /core/key_generation.py`
PRIVATE_KEY=`echo $keys | awk -F " " '{print $1}'`
PUBLIC_KEY=`echo $keys | awk -F " " '{print $2}'`

echo $PRIVATE_KEY > ${KEY}/"private_key"
echo $PUBLIC_KEY > ${KEY}/"public_key"
