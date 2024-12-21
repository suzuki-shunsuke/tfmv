# tfmv

CLI to rename Terraform resources and modules and generate moved blocks.
You can rename blocks flexibly using [Jsonnet](https://jsonnet.org).

## Install

```
go install github.com/suzuki-shunsuke/tfmv/cmd/tfmv@latest
```

## Getting Started

1. Install tfmv
1. Checkout a repository

```sh
git clone https://github.com/suzuki-shunsuke/tfmv
cd tfmv/example
```

main.tf:

```tf
resource "null_resource" "foo-prod" {}
```

Let's replace `-` with `_`.

tfmv uses Jsonnet to rename resources flexibly.
[For details of Jsonnet, please see here](#jsonnet).

tfmv.jsonnet:

```jsonnet
std.native("strings.Replace")(input.name, "-", "_", -1)
```

Run `tfmv -file tfmv.jsonnet`.
You don't need to run `terraform init`.

```sh
tfmv -file tfmv.jsonnet
```

Then a resource name is changed and `moved.tf` is created.
By default, tfmv finds *.tf on the current directory.

main.tf:

```tf
resource "null_resource" "foo_prod" {}
```

moved.tf:

```tf
moved {
  from = "null_resource.foo-prod"
  to   = "null_resource.foo_prod"
}
```

### Pass *.tf via arguments

You can also pass *.tf via arguments:

```sh
tfmv -file tfmv.jsonnet foo/aws_s3_bucket.tf foo/aws_instance.tf
```

### Change the filename for moved blocks

By default tfmv writes moved blocks to `moved.tf`.
You can change the file name via `-dest` option.

```sh
tfmv -file tfmv.jsonnet -dest moved_blocks.tf bar/main.tf
```

You can also write moved blocks to the same file with renamed resources and modules.

```sh
tfmv -file tfmv.jsonnet -dest same bar/foo.tf
```

### `-r` Recursive option

By default, tfmv finds *.tf on the current directory.
You can find files recursively using `-r` option.

```sh
tfmv -r -file tfmv.jsonnet
```

## Jsonnet

tfmv uses [Jsonnet](https://jsonnet.org) to enable you to define a custom rename logic.
You need to define Jsonnet whose input is each resource and output is a new resource name.
tfmv passes an input via External Variables.
The Jsonnet must returns a new resource name.
If the returned value is an empty string or not changed, the resource isn't renamed.

The type of an external variable `input` is as following:

```json
{
  "file": "A relative file path from the current directory to the Terraform configuration file",
  "block_type": "Either module or resource",
  "resource_type": "A resource type. e.g. null_resource. If block_type is module, resource_type is empty",
  "name": "A resource name. For example, the resource address is null_resource.foo, the name is foo."
}
```

### Native Functions

tfmv supports the following [native functions](https://pkg.go.dev/github.com/google/go-jsonnet#NativeFunction).

You can executed these functions by `std.native("{native function name}")`.

e.g.

```jsonnet
local contained = std.native("strings.Contains")("hello", "ll"); // true
```

For details, please see [Native functions](docs/native-function.md).

## LICENSE

[MIT](LICENSE)
