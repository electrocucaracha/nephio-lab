# participant

## Description
Kpt package to apply with participant specific repositories and other setup

## Usage

```bash
kpt pkg get --for-deployment https://github.com/electrocucaracha/nephio-lab/packages/participant
kpt fn render
kpt live init
kpt live apply --output table
```

This assumes the Gitea basic auth secret `gitea-personal-access-token` has
been created with username `gitea-admin` and `secret` as the password.

This will pull the package and set up the repository pointers correctly.
