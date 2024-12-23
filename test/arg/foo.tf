data "github_branch" "example-2" {
  repository = github_repository.example-1.name
  branch     = "example"
  depends_on = [
    github_repository.example-1,
  ]
}
