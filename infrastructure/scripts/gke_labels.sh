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

### GKE nodes
for cluster in $(gcloud container clusters list --project $project --format="value(name)")
do
	echo $cluster
	hasLabel=$(gcloud container clusters list --project $project \
		--filter="name=$cluster" --format="value(resourceLabels.$key)")
	region=$(gcloud container clusters list --project $project \
		--filter="name=$cluster" --format="value(location)")

	if [ "$hasLabel" = "$val" ] ; then
		echo "already $key=$val"
		if [ "$mode" = "del" ]; then
			gcloud container clusters update $cluster --project $project --region $region --remove-labels $key
			echo "------------deleted---------------"
			continue
		fi
	elif [ "$mode" = "soft" ]; then
		gcloud container clusters update $cluster --project $project --region $region --update-labels $key=$val
		echo "--------------added(soft)-----------------"
		continue
	fi

	if [ "$mode" = "hard" ]; then
		gcloud container clusters update $cluster --project $project --region $region --update-labels $key=$val
		echo "--------------added(hard)-----------------"
	fi
done

echo "gke clusters: done."
