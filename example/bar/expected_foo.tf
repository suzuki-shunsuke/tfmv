resource "null_resource" "foo_prod" {}

moved {
  from = null_resource.foo-prod
  to   = null_resource.foo_prod
}
