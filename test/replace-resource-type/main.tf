resource "aws_elasticsearch_domain" "example_1" {
  domain_name = "example-1"
}

data "aws_elasticsearch_domain" "example_2" {
  domain_name = aws_elasticsearch_domain.example_1.domain_name
}

module "aws_elasticsearch_domain" {
  source = "./mdule"
}

resource "null_resouce" "aws_elasticsearch_domain" {}
