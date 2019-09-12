#!/bin/bash
set -e

POSITIONAL=()
while [[ $# -gt 0 ]]
do
key="$1"

case $key in
    --version)
    VERSION="$2"
    shift # past argument
    shift # past value
    ;;
    --api-token)
    TOKEN="$2"
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

version="0.2.1"
if [ "${VERSION}" != "" ] ; then
    version="${VERSION}"
fi
echo " => Downloading version: ${version}"

download_url="https://github.com/bitrise-io/bitrise-add-new-project/releases/download/${version}/banp-Darwin-x86_64"
echo " => Downloading banp from (${download_url}) to (${bin_path}) ..."
curl -fL --progress-bar --output "${bin_path}" "${download_url}"
echo " => Making it executable ..."
chmod +x "${bin_path}"
echo " => Running banp ..."
echo
${bin_path} --api-token "${TOKEN}" --org "${ORG}" --public "${PUBLIC}"
