#!/bin/bash
log="./count.info"
ts=`date +"%F %T"`
echo $ts >> $log
echo "$1 add $2" >> $log 
count=1
while [[ $count -le 20 ]]
do
  echo "it is :${count}"  >> $log
  
  let count=count+1
  sleep 1
done
