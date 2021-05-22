## Kanishk's Terraform Playground

This is _not_ meant to be a legitimate Terraform replacement. It's just a place for me to mess around.

It looks like we need to run a `terraform init` before we run `terraform <anything>` if we're not providing TF with a state file, otherwise it fails to load the required providers and cannot provide the graph we need.
