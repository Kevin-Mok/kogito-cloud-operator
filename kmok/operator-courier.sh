#!/bin/bash

source ./quay-credentials.sh

export AUTH_TOKEN=$(curl -sH "Content-Type: application/json" -XPOST https://quay.io/cnr/api/v1/users/login -d '
{
    "user": {
        "username": "'"${QUAY_USERNAME}"'",
        "password": "'"${QUAY_PASSWORD}"'"
    }
}')
# | jq -r '.token'
echo $AUTH_TOKEN

export OPERATOR_DIR=build/_output/operatorhub/
export QUAY_NAMESPACE=kmok # should be different in your environment
export PACKAGE_NAME=kogito-operator
export PACKAGE_VERSION=0.11.0
export TOKEN=$AUTH_TOKEN

echo operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
# operator-courier push "$OPERATOR_DIR" "$QUAY_NAMESPACE" "$PACKAGE_NAME" "$PACKAGE_VERSION" "$TOKEN"
