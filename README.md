# tfmv

[![License](http://img.shields.io/badge/license-mit-blue.svg?style=flat-square)](https://raw.githubusercontent.com/suzuki-shunsuke/tfmv/main/LICENSE) | [Install](docs/install.md)

CLI to rename Terraform resources and modules and generate moved blocks.

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

resource "github_branch" "example" {
  repository = github_repository.example-1.name
  branch     = "example"
  depends_on = [
    github_repository.example-1
  ]
}
```

Let's replace `-` with `_`.
You need to specify either `-r` or `--jsonnet (-j)`.
In this case, let's use `-r`.
[If you need more flexible renaming, you can use Jsonnet. For details, please see here](#jsonnet).

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
index 48ce91d..e618ab1 100644
--- a/example/main.tf
+++ b/example/main.tf
@@ -1,11 +1,11 @@
-resource "github_repository" "example-1" {
+resource "github_repository" "example_1" {
   name = "example-1"
 }
 
 resource "github_branch" "example" {
-  repository = github_repository.example-1.name
+  repository = github_repository.example_1.name
   branch     = "example"
   depends_on = [
-    github_repository.example-1
+    github_repository.example_1
   ]
 }
```

moved.tf:

```tf
moved {
  from = github_repository.example-1
  to   = github_repository.example_1
}
```

### Pass *.tf via arguments

You can also pass *.tf via arguments:

```sh
tfmv -r "-/_" foo/aws_s3_bucket.tf foo/aws_instance.tf
```

tfmv supports modules too.

```sh
tfmv -r "production/prod" foo/module_foo.tf
```

### Dry Run: --dry-run

With `--dry-run`, tfmv outputs logs but doesn't rename blocks.

```sh
tfmv -r "-/_" --dry-run bar/main.tf
```

### Rename resources by regular expression

With `--regexp`, tfmv renames resources by regular expression.

```sh
tfmv --regexp '^example-(\d+)/test-$1' regexp/main.tf
```

About regular expression, please see the following document:

- https://golang.org/s/re2syntax
- https://pkg.go.dev/regexp#Regexp.ReplaceAllString

### Filter resources by regular expression

With `--include <regular expression>`, only resources matching the regular expression are renamed.

```sh
tfmv -r "-/_" --include "^aws_" bar/main.tf
```

With `--exclude <regular expression>`, only resources not matching the regular expression are renamed.

```sh
tfmv -r "-/_" --exclude "^null_resource" bar/main.tf
```

### Change the filename for moved blocks

By default tfmv writes moved blocks to `moved.tf`.
You can change the file name via `-m` option.

```sh
tfmv -r "-/_" -m moved_blocks.tf bar/main.tf
```

You can also write moved blocks to the same file with renamed resources and modules.

```sh
tfmv -r "-/_" -m same bar/foo.tf
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
