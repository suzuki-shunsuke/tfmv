# tfmv

[![DeepWiki](https://img.shields.io/badge/DeepWiki-suzuki--shunsuke%2Ftfmv-blue.svg?logo=data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAACwAAAAyCAYAAAAnWDnqAAAAAXNSR0IArs4c6QAAA05JREFUaEPtmUtyEzEQhtWTQyQLHNak2AB7ZnyXZMEjXMGeK/AIi+QuHrMnbChYY7MIh8g01fJoopFb0uhhEqqcbWTp06/uv1saEDv4O3n3dV60RfP947Mm9/SQc0ICFQgzfc4CYZoTPAswgSJCCUJUnAAoRHOAUOcATwbmVLWdGoH//PB8mnKqScAhsD0kYP3j/Yt5LPQe2KvcXmGvRHcDnpxfL2zOYJ1mFwrryWTz0advv1Ut4CJgf5uhDuDj5eUcAUoahrdY/56ebRWeraTjMt/00Sh3UDtjgHtQNHwcRGOC98BJEAEymycmYcWwOprTgcB6VZ5JK5TAJ+fXGLBm3FDAmn6oPPjR4rKCAoJCal2eAiQp2x0vxTPB3ALO2CRkwmDy5WohzBDwSEFKRwPbknEggCPB/imwrycgxX2NzoMCHhPkDwqYMr9tRcP5qNrMZHkVnOjRMWwLCcr8ohBVb1OMjxLwGCvjTikrsBOiA6fNyCrm8V1rP93iVPpwaE+gO0SsWmPiXB+jikdf6SizrT5qKasx5j8ABbHpFTx+vFXp9EnYQmLx02h1QTTrl6eDqxLnGjporxl3NL3agEvXdT0WmEost648sQOYAeJS9Q7bfUVoMGnjo4AZdUMQku50McDcMWcBPvr0SzbTAFDfvJqwLzgxwATnCgnp4wDl6Aa+Ax283gghmj+vj7feE2KBBRMW3FzOpLOADl0Isb5587h/U4gGvkt5v60Z1VLG8BhYjbzRwyQZemwAd6cCR5/XFWLYZRIMpX39AR0tjaGGiGzLVyhse5C9RKC6ai42ppWPKiBagOvaYk8lO7DajerabOZP46Lby5wKjw1HCRx7p9sVMOWGzb/vA1hwiWc6jm3MvQDTogQkiqIhJV0nBQBTU+3okKCFDy9WwferkHjtxib7t3xIUQtHxnIwtx4mpg26/HfwVNVDb4oI9RHmx5WGelRVlrtiw43zboCLaxv46AZeB3IlTkwouebTr1y2NjSpHz68WNFjHvupy3q8TFn3Hos2IAk4Ju5dCo8B3wP7VPr/FGaKiG+T+v+TQqIrOqMTL1VdWV1DdmcbO8KXBz6esmYWYKPwDL5b5FA1a0hwapHiom0r/cKaoqr+27/XcrS5UwSMbQAAAABJRU5ErkJggg==)](https://deepwiki.com/suzuki-shunsuke/tfmv) [![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/tfmv/main/LICENSE) | [Install](docs/install.md)

tfmv is a CLI to rename Terraform resources, data sources, and modules and generate moved blocks.

e.g. Replace `-` with `_`:

```sh
tfmv -r "-/_"
```

```diff
-resource "github_repository" "example-1" {
+resource "github_repository" "example_1" {
   name = "example-1"
 }
 
 data "github_branch" "example" {
-  repository = github_repository.example-1.name
+  repository = github_repository.example_1.name
   branch     = "example"
 }
```

moved.tf is created:

```tf
moved {
  from = github_repository.example-1
  to   = github_repository.example_1
}
```

## Getting Started

1. [Install tfmv](docs/install.md)
1. Checkout this repository

```sh
git clone https://github.com/suzuki-shunsuke/tfmv
cd tfmv/example
```

main.tf:

```tf
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
```

Let's replace `-` with `_`.
You must specify one of `--replace (-r)`, `--regexp`, or `--jsonnet (-j)`.
In this case, let's use `-r`.
If you need more flexible renaming, you can use [regular expression](#rename-resources-by-regular-expression) or [Jsonnet](#jsonnet). 

Run `tfmv -r "-/_"`.
You don't need to run `terraform init`.

```sh
tfmv -r "-/_"
```

Then a resource name is changed and `moved.tf` is created.
By default, tfmv finds *.tf on the current directory.

main.tf:

```diff
diff --git a/example/main.tf b/example/main.tf
index 48ef3bd..9110880 100644
--- a/example/main.tf
+++ b/example/main.tf
@@ -1,20 +1,20 @@
-resource "github_repository" "example-1" {
+resource "github_repository" "example_1" {
   name = "example-1"
 }
 
-data "github_branch" "example-2" {
-  repository = github_repository.example-1.name
+data "github_branch" "example_2" {
+  repository = github_repository.example_1.name
   branch     = "example"
   depends_on = [
-    github_repository.example-1,
-    module.example-3
+    github_repository.example_1,
+    module.example_3
   ]
 }
 
-module "example-3" {
+module "example_3" {
   source = "./foo/module"
 }
 
 output "branch_sha" {
-  value = data.github_branch.example-2.sha
+  value = data.github_branch.example_2.sha
 }
```

moved.tf:

```tf
moved {
  from = github_repository.example-1
  to   = github_repository.example_1
}

moved {
  from = module.example-3
  to   = module.example_3
}
```

### Pass *.tf via arguments

You can also pass *.tf via arguments:

```sh
tfmv -r "-/_" main.tf
```

### Dry Run: --dry-run

With `--dry-run`, tfmv outputs logs but doesn't rename blocks.

```sh
tfmv -r "-/_" --dry-run main.tf
```

### Rename resources by regular expression

With `--regexp`, tfmv renames resources by regular expression.

e.g. Remove `-prod` suffix:

```sh
tfmv --regexp '-prod$/'
```

Inside repl, `$` signs are interpreted as in [Regexp.Expand](https://pkg.go.dev/regexp#Regexp.Expand).

```sh
tfmv --regexp '^example-(\d+)/test-$1' main.tf
```

About regular expression, please see the following document:

- https://golang.org/s/re2syntax
- https://pkg.go.dev/regexp#Regexp.ReplaceAllString

### Filter resources by regular expression

With `--include <regular expression>`, only resources matching the regular expression are renamed.

e.g. Rename only AWS resources:

```sh
tfmv -r "-/_" --include "^aws_"
```

With `--exclude <regular expression>`, only resources not matching the regular expression are renamed.

e.g. Exclude AWS resources:

```sh
tfmv -r "-/_" --exclude "^aws_"
```

### Change the filename for moved blocks

By default tfmv writes moved blocks to `moved.tf`.
You can change the file name via `-m` option.

```sh
tfmv -r "-/_" -m moved_blocks.tf
```

With `-m same`, moved blocks are outputted to the same file with rename resources.

```sh
tfmv -r "-/_" -m same
```

### `--recursive (-R)` Recursive option

By default, tfmv finds *.tf on the current directory.
You can find files recursively using `-R` option.

```sh
tfmv -Rr "-/_"
```

The following directories are ignored:

- .git
- .terraform
- node_modules

## `--log-level` Log Level

You can change the log level using `--log-level` option.

```sh
tfmv -r '-/_' --log-level debug
```

## Jsonnet

`-r` is simple and useful, but sometimes you need more flexible renaming.
In that case, you can use `--jsonnet (-j)`.
[Jsonnet](https://jsonnet.org) is a powerful data configuration language.

tfmv.jsonnet (You can change the filename freely):

```jsonnet
std.native("strings.Replace")(std.extVar('input').name, "-", "_", -1)[0]
```

```sh
tfmv -j tfmv.jsonnet
```

You need to define Jsonnet whose input is each resource and output is a new resource name.
tfmv passes an input via External Variables.
You can access an input by `std.extVar('input')`.

```jsonnet
local input = std.extVar('input');
```

The type of an external variable `input` is as following:

```json
{
  "file": "A relative file path from the current directory to the Terraform configuration file",
  "block_type": "Either module or resource",
  "resource_type": "A resource type. e.g. null_resource. If block_type is module, resource_type is empty",
  "name": "A resource name. For example, the resource address is null_resource.foo, the name is foo."
}
```

e.g.

```json
{
  "file": "foo/main.tf",
  "block_type": "resource",
  "resource_type": "null_resource",
  "name": "foo"
}
```

The Jsonnet must returns a new resource name.
If the returned value is an empty string or not changed, the resource isn't renamed.

### Native Functions

tfmv supports the following [native functions](https://pkg.go.dev/github.com/google/go-jsonnet#NativeFunction).

You can executed these functions by `std.native("{native function name}")`.

e.g.

```jsonnet
std.native("strings.Replace")(input.name, "-", "_", -1)[0]
```

For details, please see [Native functions](docs/native-function.md).

## LICENSE

[MIT](LICENSE)
