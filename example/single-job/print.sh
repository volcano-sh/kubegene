#!/bin/bash
COUNTER=0
while [ $COUNTER -lt 10 ]; do
echo The counter is $COUNTER >> /xsw/kubegenetest/subjob/counter.data
let COUNTER=COUNTER+1
done
