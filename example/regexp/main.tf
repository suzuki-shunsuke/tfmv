resource "github_repository" "example-1" {
  name = "example-1"
}

resource "github_branch" "example" {
  repository = github_repository.example-1.name
  branch     = "example"
  depends_on = [
    github_repository.example-1
  ]
}