#! /bin/bash

set -e

HOST="ch-cst-20.clickhouse-system.svc.cluster.local"
PASSWORD="zswlfqvpyi"
TABLE="starexp"
CWD=$(cd "$(dirname $0)"; pwd)
SOURCE_FILE=${CWD}'/star2002-full.csv'
QUERIES_FILE=${CWD}'/queries.sql'

function download_dataset() {
	echo "Download Dataset..."
	curl -L -o star2002-full.csv.gz http://file.intra.sensetime.com/f/1ece67bb46/?raw=1
	gzip -d star2002-full.csv.gz
	echo "Download Dataset Finished."
}

function create_schema() {
	echo "Creat Table..."

	clickhouse-client -h "${HOST}" --password "${PASSWORD}" --query="
	CREATE TABLE IF NOT EXISTS ${TABLE} ON CLUSTER '{cluster}' (
	antiNucleus UInt32,
	eventFile UInt32,
	eventNumber UInt32,
	eventTime Float64,
	histFile UInt32,
	multiplicity UInt32,
	NaboveLb UInt32,
	NbelowLb UInt32,
	NLb  UInt32,
	primaryTracks UInt32,
	prodTime Float64,
	Pt  Float32,
	runNumber UInt32,
	vertexX  Float32,
	vertexY  Float32,
	vertexZ  Float32,
	eventDate Date default concat(substring(toString(floor(eventTime)), 1, 4), '-', substring(toString(floor(eventTime)), 5, 2), '-', substring(toString(floor(eventTime)), 7, 2))
	)
	ENGINE = ReplicatedMergeTree('{shard}','{replica}', eventDate, (eventNumber, eventTime, runNumber, eventFile, multiplicity), 8192);
	"

	clickhouse-client -h "${HOST}" --password "${PASSWORD}" --query="
	CREATE TABLE IF NOT EXISTS ${TABLE}_dist ON CLUSTER '{cluster}'
	AS ${TABLE}
	ENGINE = Distributed('{cluster}', default, ${TABLE}, rand());
	"

	echo "Create Table Finished."
}

function load_data() {
	echo "Load Data..."
	start=$(date +%s)
	clickhouse-client -h "${HOST}" --password "${PASSWORD}" -d default \
	    --query="INSERT INTO ${TABLE} (antiNucleus, eventFile, eventNumber, eventTime, histFile, multiplicity, NaboveLb, NbelowLb, NLb, primaryTracks, prodTime, Pt, runNumber, vertexX, vertexY, vertexZ) FORMAT CSV" \
	    < "$SOURCE_FILE"
	end=$(date +%s)
	echo "Load Data Finished."
	echo "Time Spent: "$((end-start))" Seconds."
}


function test() {
	echo "---Test Begin---"

    sed "s/{table}/${TABLE}/g" < "$QUERIES_FILE" | while read query; do
        for CONCURRENCY in 1 10 100 500 1000
        do
            echo "query: $query"
            echo "concurrency: $CONCURRENCY"
            date
            clickhouse-benchmark --concurrency=$CONCURRENCY \
            --delay=0 \
            --host="${HOST}" \
            --iterations=0 \
            --database=default \
            --cumulative \
            --password="${PASSWORD}" <<< "$query" &
            sleep 120
            kill -s INT $!
            date
            sleep 45 # Recovery from last testing
        done
    done
	echo "---Test End---"
}

#download_dataset
#create_schema
#load_data
test








