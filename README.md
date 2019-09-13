# banp (bitrise-add-new-project)

Creates new project on Bitrise.io using a local bitrise.yml or running the Bitrise scanner.

## one-liner for executing banp

Copy and Paste the above commands in your terminal on macOS and Linux.

### Create a Bitrise project under an organization

```BASH
bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/master/_scripts/run.sh) --api-token "<Bitrise personal access token>" --org "<organisation slug>" --public="<true|false>"
```

### Create a personal Bitrise project

```BASH
bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/master/_scripts/run.sh) --api-token "<Bitrise personal access token>" --public="<true|false>"
```

## Install or upgrade

```BASH
curl -fL https://github.com/bitrise-io/bitrise-add-new-project/releases/latest/download/banp-$(uname -s)-$(uname -m) > /usr/local/bin/banp
```

Then:

```BASH
chmod +x /usr/local/bin/banp
```

That's all, you're ready to call `banp`!
