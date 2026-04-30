import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getK8sChangelogTools, keepOnlyChanges } from './k8schangelog.js';
import { Config } from '../config/config.js';

const fakeChangelogContent = `<!-- BEGIN MUNGE: GENERATED_TOC -->

- [v1.33.6](#v1336)
  - [Downloads for v1.33.6](#downloads-for-v1336)
    - [Source Code](#source-code)
    - [Client Binaries](#client-binaries)
    - [Server Binaries](#server-binaries)
    - [Node Binaries](#node-binaries)
    - [Container Images](#container-images)
  - [Changelog since v1.33.5](#changelog-since-v1335)
  - [Changes by Kind](#changes-by-kind)
    - [Feature](#feature)
    - [Bug or Regression](#bug-or-regression)
    - [Other (Cleanup or Flake)](#other-cleanup-or-flake)
  - [Dependencies](#dependencies)
    - [Added](#added)
    - [Changed](#changed)
    - [Removed](#removed)
- [v1.33.5](#v1335)
  - [Downloads for v1.33.5](#downloads-for-v1335)
    - [Source Code](#source-code-1)
    - [Client Binaries](#client-binaries-1)
    - [Server Binaries](#server-binaries-1)
    - [Node Binaries](#node-binaries-1)
    - [Container Images](#container-images-1)
  - [Changelog since v1.33.4](#changelog-since-v1334)
  - [Changes by Kind](#changes-by-kind-1)
    - [Feature](#feature-1)
    - [Bug or Regression](#bug-or-regression-1)
    - [Other (Cleanup or Flake)](#other-cleanup-or-flake-1)
  - [Dependencies](#dependencies-1)
    - [Added](#added-1)
    - [Changed](#changed-1)
    - [Removed](#removed-1)

<!-- END MUNGE: GENERATED_TOC -->

# v1.33.6


## Downloads for v1.33.6



### Source Code

filename | sha512 hash
-------- | -----------
[kubernetes.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes.tar.gz) | 5bcb91f1507599d1f37f7182ee17183e995b7c1c421bc0ef103c24fe18d048a9523c273354362cc6f2bd49b4c5af0b97a27dc29f5dee34d9981a035f540a344a
[kubernetes-src.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-src.tar.gz) | 4151186f053b7e4fc9df21e810e7977e33754bee9f11a9004649a21463b49fe20278d1ab2cdfd58f9fc8420aeefea9dddaa95316aa45ee8bdb3d6ecd37aec047

### Client Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-client-darwin-amd64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-client-darwin-amd64.tar.gz) | 60056aa49cde92209248583cbea1f98b6378f808f5acdb4dbda745d9ef7e00ad4e2541081419c6d1b19703681e45230537f80b0f89d041a996f2b907d7bf71f6
[kubernetes-client-darwin-arm64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-client-darwin-arm64.tar.gz) | 5ff48911774bcdfc2c0aabf97807a1829fd1598cfff98afb240cb75bdd24e91758e95658e31af0e996e280826c098a2bb0887cf0237d3c7ad49a92bdc1d01926

### Server Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-server-linux-amd64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-server-linux-amd64.tar.gz) | 99f31e950cbff1ff30c3802983b6bc974048a4927695599ff33e1532d88c0761ea05530ef09ad974d8d165156577e0d1c414c6322ef8f5772a47e65c5270bb49
[kubernetes-server-linux-arm64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-server-linux-arm64.tar.gz) | c04f59b948cd72922f69d4093642c149e8c061fc8793120683aca292cc95bf15486a62e8d7162fe62980de03c39bc956c1b92236f94380411d33c5e65d4d8191

### Node Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-node-linux-amd64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-node-linux-amd64.tar.gz) | 9ca6ec8e0b52b9ba050479719b38da757aeca646be4a0ef1f72b2782962e7bd44c10a04688377217cbdde7b627d9ac4d9acde6c5c38a68a0678357c71d1e47ef
[kubernetes-node-linux-arm64.tar.gz](https://dl.k8s.io/v1.33.6/kubernetes-node-linux-arm64.tar.gz) | 99a50e694628fe2253aca2e0e463a7565ead046df7ab93f8dc8086cebf269fd9fcc7256e64806e487f06bb61078c7e1abe9074a743d7f192f5ec3c9fcdf55a13

### Container Images

All container images are available as manifest lists and support the described
architectures. It is also possible to pull a specific architecture directly by
adding the "-$ARCH" suffix  to the container image name.

name | architectures
---- | -------------
[registry.k8s.io/conformance:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-s390x)
[registry.k8s.io/kube-apiserver:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-s390x)
[registry.k8s.io/kube-controller-manager:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-s390x)
[registry.k8s.io/kube-proxy:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-s390x)
[registry.k8s.io/kube-scheduler:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-s390x)
[registry.k8s.io/kubectl:v1.33.6](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-s390x)

## Changelog since v1.33.5

## Changes by Kind

### Feature

- Kubernetes is now built using Go 1.24.9
  - update setcap and debian-base to bookworm-v1.0.6 ([#134613](https://github.com/kubernetes/kubernetes/pull/134613), [@cpanato](https://github.com/cpanato)) [SIG Architecture, Cloud Provider, Etcd, Release, Storage and Testing]

### Bug or Regression

- Bump system-validators to v1.9.2: remove version-specific cgroup kernel config checks to avoid false failures on cgroup v2 systems when v1-only configs are missing. ([#134086](https://github.com/kubernetes/kubernetes/pull/134086), [@pacoxu](https://github.com/pacoxu)) [SIG Cluster Lifecycle]
- Extends the nodeports scheduling plugin to consider hostPorts used by restartable init containers. ([#133390](https://github.com/kubernetes/kubernetes/pull/133390), [@SergeyKanzhelev](https://github.com/SergeyKanzhelev)) [SIG Scheduling and Testing]
- Fix Windows kube-proxy (winkernel) issue where stale RemoteEndpoints remained
  when a Deployment was referenced by multiple Services due to premature clearing
  of the terminatedEndpoints map. ([#135171](https://github.com/kubernetes/kubernetes/pull/135171), [@princepereira](https://github.com/princepereira)) [SIG Network and Windows]
- Fix Windows kube-proxy to prevent intermittent deletion of ClusterIP load balancers in HNS when internalTrafficPolicy=Local, ensuring stable service connectivity. ([#134032](https://github.com/kubernetes/kubernetes/pull/134032), [@princepereira](https://github.com/princepereira)) [SIG Network and Windows]
- Fix the bug which could result in Job status updates failing with the error:
  status.startTime: Required value: startTime cannot be removed for unsuspended job
  The error could be raised after a Job is resumed, if started and suspended previously. ([#135129](https://github.com/kubernetes/kubernetes/pull/135129), [@dejanzele](https://github.com/dejanzele)) [SIG Apps and Testing]
- Fix: The requests for a config FromClass in the status of a ResourceClaim were not referenced. ([#135105](https://github.com/kubernetes/kubernetes/pull/135105), [@LionelJouin](https://github.com/LionelJouin)) [SIG Node]
- Fixed a bug in kube-proxy nftables mode (GA as of 1.33) that fails to determine if traffic originates from a local source on the node. The issue was caused by using the wrong meta \`iif\` instead of \`iifname\` for name based matches. ([#134099](https://github.com/kubernetes/kubernetes/pull/134099), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Fixed a bug in kube-proxy nftables mode (GA as of 1.33) that fails to determine if traffic originates from a local source on the node. The issue was caused by using the wrong meta \`iif\` instead of \`iifname\` for name based matches. ([#134117](https://github.com/kubernetes/kubernetes/pull/134117), [@jack4it](https://github.com/jack4it)) [SIG Network]
- Fixed a startup probe race condition that caused main containers to remain stuck in "Initializing" state when sidecar containers with startup probes failed initially but succeeded on restart in pods with restartPolicy=Never. ([#134801](https://github.com/kubernetes/kubernetes/pull/134801), [@yuanwang04](https://github.com/yuanwang04)) [SIG Node and Testing]
- Fixed race-condition in service allocation logic which leads to spurious IPAddressWrongReference warnings impacting performance ([#133954](https://github.com/kubernetes/kubernetes/pull/133954), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Fixes spammy incorrect "Ignoring same-zone topology hints for service since no hints were provided for zone" messages in the kube-proxy logs. ([#133527](https://github.com/kubernetes/kubernetes/pull/133527), [@danwinship](https://github.com/danwinship)) [SIG Network]
- Kube-controller-manager: Fixes a 1.33 regression in daemonset handling of orphaned pods ([#134652](https://github.com/kubernetes/kubernetes/pull/134652), [@liggitt](https://github.com/liggitt)) [SIG Apps]
- Kube-controller-manager: Resolves potential issues handling pods with incorrect uids in their ownerReference ([#134662](https://github.com/kubernetes/kubernetes/pull/134662), [@liggitt](https://github.com/liggitt)) [SIG Apps]
- Kube-proxy in nftables mode now allows pods on nodes without local service endpoints to access LoadBalancer Service ExternalIPs (with \`externalTrafficPolicy: Local\`). Previously, such traffic was dropped. This change brings nftables mode in line with iptables and IPVS modes, allowing traffic to be forwarded to available endpoints elsewhere in the cluster. ([#133969](https://github.com/kubernetes/kubernetes/pull/133969), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Kubeadm: avoid panicing if the user has malformed the kubeconfig in the cluster-info config map to not include a valid current context. Include proper validation at the appropriate locations and throw errors instead. ([#134724](https://github.com/kubernetes/kubernetes/pull/134724), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: ensured waiting for apiserver uses a local client that doesn't reach to the control plane endpoint and instead reaches directly to the local API server endpoint. ([#134269](https://github.com/kubernetes/kubernetes/pull/134269), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: fixed a bug where the node registration information for a given node was not fetched correctly during "kubeadm upgrade node" and the node name can end up being incorrect in cases where the node name is not the same as the host name. ([#134363](https://github.com/kubernetes/kubernetes/pull/134363), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: fixes a preflight check that can fail hostname construction in IPV6 setups ([#134590](https://github.com/kubernetes/kubernetes/pull/134590), [@liggitt](https://github.com/liggitt)) [SIG API Machinery, Auth, Cloud Provider, Cluster Lifecycle and Testing]
- Reduce event spam during volume operation errors in Portworx in-tree driver ([#135192](https://github.com/kubernetes/kubernetes/pull/135192), [@gohilankit](https://github.com/gohilankit)) [SIG Storage]

### Other (Cleanup or Flake)

- Kubeadm: updated the supported etcd version to v3.5.24 for the skewed control plane version v1.33. ([#135018](https://github.com/kubernetes/kubernetes/pull/135018), [@hakman](https://github.com/hakman)) [SIG Cloud Provider, Cluster Lifecycle and Etcd]
- Kubernetes is now built using Go 1.24.7 ([#134197](https://github.com/kubernetes/kubernetes/pull/134197), [@cpanato](https://github.com/cpanato)) [SIG Release and Testing]
- The test is intended to verify pod scheduling with an anti-affinity scenario, but it uses the wrong pod template. 
  This affects functional correctness. ([#134262](https://github.com/kubernetes/kubernetes/pull/134262), [@sats-23](https://github.com/sats-23)) [SIG Testing]

## Dependencies

### Added
_Nothing has changed._

### Changed
- k8s.io/system-validators: v1.9.1 → v1.9.2

### Removed
_Nothing has changed._




# v1.33.5


## Downloads for v1.33.5



### Source Code

filename | sha512 hash
-------- | -----------
[kubernetes.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes.tar.gz) | 7cf4e067ea5882db3d0f5e2f15a27a670ddb4d0a9ac58e26ac554e1d60b57e2c09525e64776ddad5e167d942a7020f61bba8c1c54f7a8b75b9509c5aca6898c0
[kubernetes-src.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-src.tar.gz) | ae9e888eec40a41ff0ef22a98e0a024396af375dd1ad55ca163ecde14bfa1fd3c17ba31d7c60e6d4af57aface36cbf922175ce9f7588a89497da4d252eca6623

### Client Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-client-darwin-amd64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-client-darwin-amd64.tar.gz) | d781af21ab4dc79df263162b0692cebf088c7cc75683a528d9238bc1a7d5258b51e7a7ec597d85567c09e806ff3876d0d124751545a75fa032defc8cfacd2686
[kubernetes-client-darwin-arm64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-client-darwin-arm64.tar.gz) | 102719580db30fe34f2f15251bfaa99ba035eeb54532fd201d11d29c1dabcb793d4d2b980302cf989c201178c071da7abcb8e6ee83fbf338d7d3a4b6e633441a

### Server Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-server-linux-amd64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-server-linux-amd64.tar.gz) | b48edac93e28565aa44c431099004de38a3ea896e1ebc4ecfe9ebe3eb712c5fd28ad721ef287ff970b0ad1b83f5dfbaeaca8258ede36e795b08803381a322e19
[kubernetes-server-linux-arm64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-server-linux-arm64.tar.gz) | d8e5992b240b1f174bcd84e1de5a901605d93e7e343551730477edd1c4066afcf834be174106135794a63bcf869b99aa500e17cae4238f0be68cb054eb6c6729

### Node Binaries

filename | sha512 hash
-------- | -----------
[kubernetes-node-linux-amd64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-node-linux-amd64.tar.gz) | f5e25af84feeec774522e727a05372442b3bd61ccf27da7d7c5fb08aadab5d1203ad3403cd01d4d13f550819ab7527e8b90af9903a221724e6d21baf9e590228
[kubernetes-node-linux-arm64.tar.gz](https://dl.k8s.io/v1.33.5/kubernetes-node-linux-arm64.tar.gz) | dbf946b03b5a9d39edd58f29d51a2f457fd5b85ab4e9db567ca12e69cad832d049ce4d79a404df381ea3795af1a7fdd72a887e9a2a1ec3f9ab67ab50837ad21e

### Container Images

All container images are available as manifest lists and support the described
architectures. It is also possible to pull a specific architecture directly by
adding the "-$ARCH" suffix  to the container image name.

name | architectures
---- | -------------
[registry.k8s.io/conformance:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/conformance-s390x)
[registry.k8s.io/kube-apiserver:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-apiserver-s390x)
[registry.k8s.io/kube-controller-manager:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-controller-manager-s390x)
[registry.k8s.io/kube-proxy:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-proxy-s390x)
[registry.k8s.io/kube-scheduler:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kube-scheduler-s390x)
[registry.k8s.io/kubectl:v1.33.5](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl) | [amd64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-amd64), [arm64](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-arm64), [ppc64le](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-ppc64le), [s390x](https://console.cloud.google.com/artifacts/docker/k8s-artifacts-prod/southamerica-east1/images/kubectl-s390x)

## Changelog since v1.33.4

## Changes by Kind

### Feature

- Kubernetes is now built using Go 1.24.6 ([#133522](https://github.com/kubernetes/kubernetes/pull/133522), [@cpanato](https://github.com/cpanato)) [SIG Release and Testing]

### Bug or Regression

- Adjusted the conformance test for the ServiceCIDR API to not test Patch/Update,
  since they are listed as ineligible_endpoints for conformance. ([#133642](https://github.com/kubernetes/kubernetes/pull/133642), [@danwinship](https://github.com/danwinship)) [SIG Network and Testing]
- Fixed SELinux warning controller not emitting events on some SELinux label conflicts. ([#133746](https://github.com/kubernetes/kubernetes/pull/133746), [@jsafrane](https://github.com/jsafrane)) [SIG Apps, Storage and Testing]
- Kubeadm: fixed bug where v1beta3's ClusterConfiguration.APIServer.TimeoutForControlPlane is not respected in newer versions of kubeadm where v1beta4 is the default. ([#133754](https://github.com/kubernetes/kubernetes/pull/133754), [@HirazawaUi](https://github.com/HirazawaUi)) [SIG Cluster Lifecycle]

### Other (Cleanup or Flake)

- Masked off access to Linux thermal interrupt info in \`/proc\` and \`/sys\`. ([#132985](https://github.com/kubernetes/kubernetes/pull/132985), [@saschagrunert](https://github.com/saschagrunert)) [SIG Node]
`

const expectedProcessedContent = `# v1.33.6


## Changelog since v1.33.5

## Changes by Kind

### Feature

- Kubernetes is now built using Go 1.24.9
  - update setcap and debian-base to bookworm-v1.0.6 ([#134613](https://github.com/kubernetes/kubernetes/pull/134613), [@cpanato](https://github.com/cpanato)) [SIG Architecture, Cloud Provider, Etcd, Release, Storage and Testing]

### Bug or Regression

- Bump system-validators to v1.9.2: remove version-specific cgroup kernel config checks to avoid false failures on cgroup v2 systems when v1-only configs are missing. ([#134086](https://github.com/kubernetes/kubernetes/pull/134086), [@pacoxu](https://github.com/pacoxu)) [SIG Cluster Lifecycle]
- Extends the nodeports scheduling plugin to consider hostPorts used by restartable init containers. ([#133390](https://github.com/kubernetes/kubernetes/pull/133390), [@SergeyKanzhelev](https://github.com/SergeyKanzhelev)) [SIG Scheduling and Testing]
- Fix Windows kube-proxy (winkernel) issue where stale RemoteEndpoints remained
  when a Deployment was referenced by multiple Services due to premature clearing
  of the terminatedEndpoints map. ([#135171](https://github.com/kubernetes/kubernetes/pull/135171), [@princepereira](https://github.com/princepereira)) [SIG Network and Windows]
- Fix Windows kube-proxy to prevent intermittent deletion of ClusterIP load balancers in HNS when internalTrafficPolicy=Local, ensuring stable service connectivity. ([#134032](https://github.com/kubernetes/kubernetes/pull/134032), [@princepereira](https://github.com/princepereira)) [SIG Network and Windows]
- Fix the bug which could result in Job status updates failing with the error:
  status.startTime: Required value: startTime cannot be removed for unsuspended job
  The error could be raised after a Job is resumed, if started and suspended previously. ([#135129](https://github.com/kubernetes/kubernetes/pull/135129), [@dejanzele](https://github.com/dejanzele)) [SIG Apps and Testing]
- Fix: The requests for a config FromClass in the status of a ResourceClaim were not referenced. ([#135105](https://github.com/kubernetes/kubernetes/pull/135105), [@LionelJouin](https://github.com/LionelJouin)) [SIG Node]
- Fixed a bug in kube-proxy nftables mode (GA as of 1.33) that fails to determine if traffic originates from a local source on the node. The issue was caused by using the wrong meta \`iif\` instead of \`iifname\` for name based matches. ([#134099](https://github.com/kubernetes/kubernetes/pull/134099), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Fixed a bug in kube-proxy nftables mode (GA as of 1.33) that fails to determine if traffic originates from a local source on the node. The issue was caused by using the wrong meta \`iif\` instead of \`iifname\` for name based matches. ([#134117](https://github.com/kubernetes/kubernetes/pull/134117), [@jack4it](https://github.com/jack4it)) [SIG Network]
- Fixed a startup probe race condition that caused main containers to remain stuck in "Initializing" state when sidecar containers with startup probes failed initially but succeeded on restart in pods with restartPolicy=Never. ([#134801](https://github.com/kubernetes/kubernetes/pull/134801), [@yuanwang04](https://github.com/yuanwang04)) [SIG Node and Testing]
- Fixed race-condition in service allocation logic which leads to spurious IPAddressWrongReference warnings impacting performance ([#133954](https://github.com/kubernetes/kubernetes/pull/133954), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Fixes spammy incorrect "Ignoring same-zone topology hints for service since no hints were provided for zone" messages in the kube-proxy logs. ([#133527](https://github.com/kubernetes/kubernetes/pull/133527), [@danwinship](https://github.com/danwinship)) [SIG Network]
- Kube-controller-manager: Fixes a 1.33 regression in daemonset handling of orphaned pods ([#134652](https://github.com/kubernetes/kubernetes/pull/134652), [@liggitt](https://github.com/liggitt)) [SIG Apps]
- Kube-controller-manager: Resolves potential issues handling pods with incorrect uids in their ownerReference ([#134662](https://github.com/kubernetes/kubernetes/pull/134662), [@liggitt](https://github.com/liggitt)) [SIG Apps]
- Kube-proxy in nftables mode now allows pods on nodes without local service endpoints to access LoadBalancer Service ExternalIPs (with \`externalTrafficPolicy: Local\`). Previously, such traffic was dropped. This change brings nftables mode in line with iptables and IPVS modes, allowing traffic to be forwarded to available endpoints elsewhere in the cluster. ([#133969](https://github.com/kubernetes/kubernetes/pull/133969), [@aroradaman](https://github.com/aroradaman)) [SIG Network]
- Kubeadm: avoid panicing if the user has malformed the kubeconfig in the cluster-info config map to not include a valid current context. Include proper validation at the appropriate locations and throw errors instead. ([#134724](https://github.com/kubernetes/kubernetes/pull/134724), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: ensured waiting for apiserver uses a local client that doesn't reach to the control plane endpoint and instead reaches directly to the local API server endpoint. ([#134269](https://github.com/kubernetes/kubernetes/pull/134269), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: fixed a bug where the node registration information for a given node was not fetched correctly during "kubeadm upgrade node" and the node name can end up being incorrect in cases where the node name is not the same as the host name. ([#134363](https://github.com/kubernetes/kubernetes/pull/134363), [@neolit123](https://github.com/neolit123)) [SIG Cluster Lifecycle]
- Kubeadm: fixes a preflight check that can fail hostname construction in IPV6 setups ([#134590](https://github.com/kubernetes/kubernetes/pull/134590), [@liggitt](https://github.com/liggitt)) [SIG API Machinery, Auth, Cloud Provider, Cluster Lifecycle and Testing]
- Reduce event spam during volume operation errors in Portworx in-tree driver ([#135192](https://github.com/kubernetes/kubernetes/pull/135192), [@gohilankit](https://github.com/gohilankit)) [SIG Storage]

### Other (Cleanup or Flake)

- Kubeadm: updated the supported etcd version to v3.5.24 for the skewed control plane version v1.33. ([#135018](https://github.com/kubernetes/kubernetes/pull/135018), [@hakman](https://github.com/hakman)) [SIG Cloud Provider, Cluster Lifecycle and Etcd]
- Kubernetes is now built using Go 1.24.7 ([#134197](https://github.com/kubernetes/kubernetes/pull/134197), [@cpanato](https://github.com/cpanato)) [SIG Release and Testing]
- The test is intended to verify pod scheduling with an anti-affinity scenario, but it uses the wrong pod template. 
  This affects functional correctness. ([#134262](https://github.com/kubernetes/kubernetes/pull/134262), [@sats-23](https://github.com/sats-23)) [SIG Testing]

# v1.33.5


## Changelog since v1.33.4

## Changes by Kind

### Feature

- Kubernetes is now built using Go 1.24.6 ([#133522](https://github.com/kubernetes/kubernetes/pull/133522), [@cpanato](https://github.com/cpanato)) [SIG Release and Testing]

### Bug or Regression

- Adjusted the conformance test for the ServiceCIDR API to not test Patch/Update,
  since they are listed as ineligible_endpoints for conformance. ([#133642](https://github.com/kubernetes/kubernetes/pull/133642), [@danwinship](https://github.com/danwinship)) [SIG Network and Testing]
- Fixed SELinux warning controller not emitting events on some SELinux label conflicts. ([#133746](https://github.com/kubernetes/kubernetes/pull/133746), [@jsafrane](https://github.com/jsafrane)) [SIG Apps, Storage and Testing]
- Kubeadm: fixed bug where v1beta3's ClusterConfiguration.APIServer.TimeoutForControlPlane is not respected in newer versions of kubeadm where v1beta4 is the default. ([#133754](https://github.com/kubernetes/kubernetes/pull/133754), [@HirazawaUi](https://github.com/HirazawaUi)) [SIG Cluster Lifecycle]

### Other (Cleanup or Flake)

- Masked off access to Linux thermal interrupt info in \`/proc\` and \`/sys\`. ([#132985](https://github.com/kubernetes/kubernetes/pull/132985), [@saschagrunert](https://github.com/saschagrunert)) [SIG Node]`;

describe('k8schangelog', () => {
  const config = new Config('1.0.0');
  const tools = getK8sChangelogTools(config);
  const tool = tools.find(t => t.name === 'get_k8s_changelog');

  if (!tool) throw new Error('get_k8s_changelog tool not found');

  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('should fetch and process changelog successfully', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      text: async () => fakeChangelogContent,
    }));

    const result = await tool.handler({ KubernetesMinorVersion: "1.31" });
    expect(result.content[0].text.trim()).toBe(expectedProcessedContent.trim());
  });

  it('should throw error for invalid version', async () => {
    await expect(tool.handler({ KubernetesMinorVersion: "1.31.5" })).rejects.toThrow("invalid kubernetes minor version: 1.31.5");
  });

  it('should handle 404 from GitHub', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
    }));

    await expect(tool.handler({ KubernetesMinorVersion: "1.32" })).rejects.toThrow("failed to get changelog with status code: 404");
  });

  it('should test keepOnlyChanges directly', () => {
    const input = `
This is some text before the first version.
It should be ignored.

# v1.2.3

## Downloads for v1.2.3

- binary 1
- binary 2

## Changelog since v1.2.2

## Changes by Kind

### Changes of Kind A
- A change.

### Changes of Kind B
- B change.

## Dependencies
- Some dependency 1
- Some dependency 2

# v1.2.2
`;
    const expected = `# v1.2.3

## Changelog since v1.2.2

## Changes by Kind

### Changes of Kind A
- A change.

### Changes of Kind B
- B change.

# v1.2.2
`;
    expect(keepOnlyChanges(input).trim()).toBe(expected.trim());
  });
});
