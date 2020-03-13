#!/bin/bash
cp ../../spush.go .
cp -R ../../tools .
cd bin
find . -type f | xargs -I{} dos2unix {}
