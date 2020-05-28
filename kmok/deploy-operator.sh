#!/bin/bash

oc create -f ../deploy/service_account.yaml
oc create -f ../deploy/role.yaml
oc create -f ../deploy/role_binding.yaml
oc create -f ../deploy/operator.yaml
