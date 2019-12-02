#!/bin/bash
set -e

host={$1-:`hostname -f`}
CWD=$(cd "$(dirname $0)"; pwd)
start=$(date +%s)
sourceFile=${CWD}'/star2002-full.csv';

clickhouse-client -h "${host}" --query="CREATE DATABASE IF NOT EXISTS test";
cat ${CWD}/merge-tree.sql | clickhouse-client -h "${host}" -mn
echo "start load data";

cat $sourceFile | clickhouse-client -h ${host} -d test --query="INSERT INTO starexp (antiNucleus, eventFile, eventNumber, eventTime, histFile, multiplicity, NaboveLb, NbelowLb, NLb, primaryTracks, prodTime, Pt, runNumber, vertexX, vertexY, vertexZ) FORMAT CSV"
echo "finish load"

end=$(date +%s)
echo "Complete in "$((end-start))" seconds."
