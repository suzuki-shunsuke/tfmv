resource "github_repository" "example-1" {
  name = "example-1"
}

output "branch_sha" {
  value = data.github_branch.example-2.sha
}
