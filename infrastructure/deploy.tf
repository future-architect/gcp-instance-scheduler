provider "google" {
  project = "${lookup(var.project_name, "${terraform.env}")}"
  region  = "asia-northeast1"
}

data "archive_file" "function_zip" {
  type        = "zip"
  source_dir  = "${path.module}/../src"
  output_path = "${path.module}/files/stop_instance.zip"
}

resource "google_storage_bucket" "scheduler_bucket" {
  name          = "${lookup(var.project_name, "${terraform.env}")}-scheduler-bucket"
  project       = "${lookup(var.project_name, "${terraform.env}")}"
  location      = "asia"
  force_destroy = true
}

resource "google_storage_bucket_object" "stop_instance_zip" {
  name   = "stop_instance.zip"
  bucket = "${google_storage_bucket.scheduler_bucket.name}"
  source = "${path.module}/files/stop_instance.zip"
}

resource "google_pubsub_topic" "stop_event" {
  name    = "stop-instance-event"
  project = "${lookup(var.project_name, "${terraform.env}")}"
}

resource "google_cloudfunctions_function" "stop_instance" {
  name        = "StopInstance"
  project     = "${lookup(var.project_name, "${terraform.env}")}"
  region      = "asia-northeast1"
  runtime     = "go111"
  entry_point = "StopInstance"

  source_archive_bucket = "${google_storage_bucket.scheduler_bucket.name}"
  source_archive_object = "${google_storage_bucket_object.stop_instance_zip.name}"

  event_trigger = {
    event_type = "providers/cloud.pubsub/eventTypes/topic.publish"
    resource   = "${google_pubsub_topic.stop_event.name}"
  }
}

resource "google_cloud_scheduler_job" "shutdown-scheduler" {
  name        = "shutdown-workday"
  project     = "${lookup(var.project_name, "${terraform.env}")}"
  schedule    = "0 22 * * *"
  description = "automatically stop instances"
  time_zone   = "Asia/Tokyo"

  pubsub_target {
    topic_name = "${google_pubsub_topic.stop_event.id}"
    data       = "${base64encode("{\"command\":\"stop\"}")}"
  }
}
