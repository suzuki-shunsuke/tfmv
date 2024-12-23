resource "github_repository" "example-1" {
  name = "example-1"
}

data "github_branch" "example-2" {
  repository = github_repository.example-1.name
  branch     = "example"
  depends_on = [
    github_repository.example-1,
  ]
}

output "branch_sha" {
  value = data.github_branch.example-2.sha
}
