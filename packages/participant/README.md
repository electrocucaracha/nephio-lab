# participant

## Description
Kpt package to apply with participant specific repositories and other setup

## Usage

When you fetch the package, you should give it the name of the participant. So,
if the participant is 'workshopper', then:

```bash
PARTICIPANT=nephio-poc-001 # Replace 'workshopper' with the participant name provided to you

kpt pkg get --for-deployment https://github.com/electrocucaracha/nephio-lab/packages/participant $PARTICIPANT
kpt fn render $PARTICIPANT
kpt live init $PARTICIPANT
kpt live apply $PARTICIPANT --output table
```

This assumes the Gitea basic auth secret `gitea-personal-access-token` has
been created with username `gitea-admin` and `secret` as the password.

This will pull the package and set up the repository pointers correctly.
