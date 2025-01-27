#!/bin/bash

set -euxo pipefail

function eventually {
  for _ in $(seq 1 3); do
    "$@" && return 0
  done
  return 1
}

# usage: ./deploy-fleet.sh ghcr.io/rancher/fleet:sha-49f6f81 ghcr.io/rancher/fleet-agent:1h
if [ $# -ge 2 ] && [ -n "$1" ] && [ -n "$2" ]; then
  fleetRepo="${1%:*}"
  fleetTag="${1#*:}"
  agentRepo="${2%:*}"
  agentTag="${2#*:}"
else
  fleetRepo="rancher/fleet"
  fleetTag="dev"
  agentRepo="rancher/fleet-agent"
  agentTag="dev"
fi

eventually helm upgrade --install fleet-crd charts/fleet-crd \
  --atomic \
  -n cattle-fleet-system \
  --create-namespace
eventually helm upgrade --install fleet charts/fleet \
  --atomic \
  -n cattle-fleet-system \
  --create-namespace \
  --set image.repository="$fleetRepo" \
  --set image.tag="$fleetTag" \
  --set agentImage.repository="$agentRepo" \
  --set agentImage.tag="$agentTag" \
  --set agentImage.imagePullPolicy=IfNotPresent

# wait for controller and agent rollout
kubectl -n cattle-fleet-system rollout status deploy/fleet-controller
{ grep -E -q -m 1 "fleet-agent-local.*1/1"; kill $!; } < <(kubectl get bundles -n fleet-local -w)
kubectl -n cattle-fleet-system rollout status statefulset/fleet-agent

# label local cluster
kubectl patch clusters.fleet.cattle.io -n fleet-local local --type=json -p '[{"op": "add", "path": "/metadata/labels/management.cattle.io~1cluster-display-name", "value": "local" }]'
