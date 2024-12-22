resource "github_branch" "example" {
  repository = github_repository.example_1.name
  branch     = "example"
  depends_on = [
    github_repository.example_1
  ]
}
