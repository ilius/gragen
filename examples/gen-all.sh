#!/bin/bash
MYPATH="`realpath \"$0\"`"
MYDIR="`dirname \"$MYPATH\"`"
find "$MYDIR" -name gen.sh -print -execdir './gen.sh' \;
