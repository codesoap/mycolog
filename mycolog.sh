#!/bin/bash
echo "binary location debug information"
export MYCOLOG_BIN_LOC=$(which mycolog)
echo "running mycolog from: $MYCOLOG_BIN_LOC"
mycolog
