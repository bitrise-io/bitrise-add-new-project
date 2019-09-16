#!/bin/bash
set -e

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --api-token)
    TOKEN="$2"
    shift # past argument
    shift # past value
    ;;
    --personal)
    PERSONAL="$2"
    shift # past argument
    shift # past value
    ;;
    --org)
    ORG="$2"
    shift # past argument
    shift # past value
    ;;
    --public)
    PUBLIC="$2"
    shift # past argument
    ;;
    *)    # unknown option
    POSITIONAL+=("$1") # save it in an array for later
    shift # past argument
    ;;
esac
done
set -- "${POSITIONAL[@]}" # restore positional parameters

echo " => Creating a temporary directory for banp (bitrise-add-new-project) ..."
temp_dir="$(mktemp -d -t banpXXXXXX)"
bin_path="${temp_dir}/banp"

download_url="https://github.com/bitrise-io/bitrise-add-new-project/releases/latest/download/banp-$(uname -s)-$(uname -m)"
echo " => Downloading banp from (${download_url}) to (${bin_path}) ..."
curl -fL --progress-bar --output "${bin_path}" "${download_url}"
echo " => Making it executable ..."
chmod +x "${bin_path}"
echo " => Running banp ..."
echo

if [ -z "$ORG" ]
then
    if [ -z "$PERSONAL" ]
    then
        ${bin_path} --api-token "${TOKEN}" --public="${PUBLIC}"
    else
        ${bin_path} --api-token "${TOKEN}" --public="${PUBLIC}" --personal="${PERSONAL}"
    fi
else
    ${bin_path} --api-token "${TOKEN}" --org "${ORG}" --public="${PUBLIC}"
fi
