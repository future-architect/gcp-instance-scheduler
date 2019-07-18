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

### compute instances
for instance in $(gcloud compute instances list --project $project --format="value(name)")
do
	echo $instance
	hasLabel=$(gcloud compute instances list --project $project \
		--filter="$instance" --format="value(labels.$key)")
	zone=$(gcloud compute instances list --project $project --filter="$instance" --format="value(zone)")
	if [ "$hasLabel" = "$val" ]; then
		echo "already $key=$val"
		if [ "$mode" = "del" ]; then
			gcloud compute instances update $instance --project $project --zone $zone --remove-labels=$key
			echo "------------deleted---------------"
			continue
		fi
	elif [ "$mode" = "soft" ]; then
		gcloud compute instances add-labels $instance --project $project --zone $zone --labels=$key=$val
		echo "--------------added(soft)-----------------"
		continue
	fi

	if [ "$mode" = "hard" ]; then
		gcloud compute instances add-labels $instance --project $project --zone $zone --labels=$key=$val
		echo "--------------added(hard)-----------------"
	fi
done

echo "compute instance: done."

