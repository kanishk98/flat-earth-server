resource "aws_s3_bucket" "test_bucket_resource_name" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = {
    Name        = "My bucket"
    Environment = "Dev"
  }
}

data "aws_vpc" "foo" {
  vpc_id = "a_vpc_id"
}

data "aws_vpc_endpoint" "s3_vpce" {
  vpc_id = aws_vpc.foo.id
  service_name = "vpce.service-endpoint.com"
}