provider "aws" {
  region = "us-east-1"
}

resource "random_string" "obj_suffix" {
  length  = 4
  special = false
  upper   = false
}

resource "aws_s3_bucket" "obj_storage_aws" {
  bucket = "multy-test-${random_string.obj_suffix.result}"
}

resource "aws_s3_object" "file1_public_aws" {
  bucket         = aws_s3_bucket.obj_storage_aws.id
  key            = "hello_world"
  acl            = "public-read"
  content_base64 = base64encode("<h1>hello world from aws</h1>")
  content_type   = "text/html"
}

output "aws_object_url" {
  value = "https://${aws_s3_bucket.obj_storage_aws.bucket}.s3.amazonaws.com/${aws_s3_object.file1_public_aws.key}"
}