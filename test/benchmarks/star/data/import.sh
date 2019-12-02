#!/bin/bash
set -e

CWD=$(cd $(dirname $0); pwd)
start=$(date +%s)
sourceFile=${CWD}'/star2002-full.csv';

echo "start load f";
cat $sourceFile | clickhouse-client --query="INSERT INTO starexp (antiNucleus, eventFile, eventNumber, eventTime, histFile, multiplicity, NaboveLb, NbelowLb, NLb, primaryTracks, prodTime, Pt, runNumber, vertexX, vertexY, vertexZ) FORMAT CSV"
echo "finish load"

end=$(date +%s)
echo "Complete in "$((end-start))" seconds."
