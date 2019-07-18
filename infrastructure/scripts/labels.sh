#!/bin/bash

project=$1
mode=$2
key=$3
val=$4

## validate parameters
if [ -z "$project" ] || [ -z "$mode" ] || [ -z "$key" ] || [ -z "$val" ] ; then
	echo "Usage: ./labels.sh [project_id] [mode:(hard, soft, del)] [label_key] [label_value]"
	exit 1
elif [ "$mode" != "hard" ] && [ "$mode" != "soft" ] && [ "$mode" != "del" ] ; then
	echo "Invalid mode: $mode"
	exit 1
fi

./gce_label $project $mode $key $val
./sql_label $project $mode $key $val
./gke_label $project $mode $key $val

echo "finish."
