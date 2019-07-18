resource "google_storage_bucket" "terraform-state-store" {
  name          = "${terraform.env}-terraform-state-store"
  location      = "asia"
  storage_class = "STANDARD"
}

resource "google_storage_bucket_acl" "remote-acl" {
  bucket         = "${google_storage_bucket.terraform-state-store.name}"
  predefined_acl = "private"
}
