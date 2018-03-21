#!/bin/bash
SCRIPTPATH=$( cd "$(dirname "$0")" ; pwd -P )
ROOTDIR=$SCRIPTPATH/..
cd $ROOTDIR

ret=0
for fn in $(find ${ROOTDIR} -name '*.go' | grep -v vendor | grep -v testdata); do
  if [[ $fn == *.pb.go ]];then
    continue
  fi

  head -20 $fn | grep "auto\-generated" > /dev/null
  if [[ $? -eq 0 ]]; then
          continue
  fi

  head -20 $fn | grep "DO NOT EDIT" > /dev/null
  if [[ $? -eq 0 ]]; then
          continue
  fi

  head -20 $fn | grep "Apache License, Version 2" > /dev/null
  if [[ $? -ne 0 ]]; then
    echo "${fn} missing license"
    ret=$(($ret+1))
  fi

  head -20 $fn | grep Copyright > /dev/null
  if [[ $? -ne 0 ]]; then
    echo "${fn} missing Copyright"
    ret=$(($ret+1))
  fi
done

exit $ret
