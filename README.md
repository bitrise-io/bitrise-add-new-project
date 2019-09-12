# bitrise-add-new-project

Creates new project on Bitrise.io using a local existing bitrise.yml or running the Bitrise scanner.

## one-liner

__Create a Bitrise project under an organization___

`bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/installer/_scripts/install.sh) --version "0.2.1" --api-token "<Bitrise personal access token>" --org "<organisation slug>" --public "<true|false>"`

__Create a personal Bitrise project__

`bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/installer/_scripts/install.sh) --version "0.2.1" --api-token "<Bitrise personal access token>" --public "<true|false>"`