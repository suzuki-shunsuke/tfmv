resource "github_repository" "example-1" {
  name = "example-1"
}

data "github_branch" "example-2" {
  repository = github_repository.example-1.name
  branch     = "example"
  depends_on = [
    github_repository.example-1,
    module.example-3
  ]
}

module "example-3" {
  source = "./foo/module"
}

output "branch_sha" {
  value = data.github_branch.example-2.sha
}
