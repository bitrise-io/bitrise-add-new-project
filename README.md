# bitrise-add-new-project

Creates new project on Bitrise.io using a local existing bitrise.yml or running the Bitrise scanner.

## one-liner

### Create a Bitrise project under an organization

`bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/installer/_scripts/install.sh) --version "0.2.1" --api-token "<Bitrise personal access token>" --org "<organisation slug>" --public "<true|false>"`

### Create a personal Bitrise project

`bash <(curl -sfL https://raw.githubusercontent.com/bitrise-io/bitrise-add-new-project/installer/_scripts/install.sh) --version "0.2.1" --api-token "<Bitrise personal access token>" --public "<true|false>"`