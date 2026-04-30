import { describe, it, expect } from 'vitest';
import { extractReleaseNotesRelevantForUpgrade } from './gkereleasenotes.js';

const fullNotes = `
November 14, 2025

      Feature
      In GKE version 1.35.2-gke.3040000 and later, GKE rejects
anonymous requests to cluster endpoints by default for all new Autopilot or
Standard clusters.

November 11, 2025

      Feature
      The N4D machine family is now Generally Available (GA) for
Standard and Autopilot mode. For cluster autoscaler, node pool auto-creation, and Autopilot mode use
GKE version 1.34.1-gke.2037000 and later.

November 07, 2025

      Feature
      In GKE version 1.34.1-gke.2037001 and later, the
GKE logging agent in your clusters can process logs up to two
times faster.
      Feature
      In version 1.34.1-gke.1829001 and later, GKE can
auto-create multiple
node pools concurrently.

October 31, 2025

      Feature
      The Multi-Cluster Services (MCS) feature has been updated with a finalizer to
more effectively prevent potential resource leaks.

October 28, 2025

      Feature
      You can use the G4 VM, powered by NVIDIA's RTX PRO 6000 GPUs, with
GKE Autopilot in version 1.34.1-gke.1829001 or later.
      Feature
      Autoscaled blue-green upgrades are available in Preview for
GKE Standard node pools.

October 21, 2025

      Feature
      The G4 VM is generally available on GKE.
For GKE Standard, use GKE version
1.34.0-gke.1662000 or later.

October 17, 2025

      Issue
      Don't use GKE version 1.34.1-gke.1431000 or later when creating
or upgrading node pools with the a3-highgpu-8g machine type.

October 14, 2025

      Issue
      In GKE versions 1.32.4-gke.1029000 and later, MountVolume calls
for network file system (NFS) volumes might fail.

October 09, 2025

      Feature
      In GKE version 1.33.4-gke.1055000 or later, you can control
how external traffic reaches your Services on GKE clusters by
using Network Service Tiers.
      Feature
      In GKE version 1.30.3-gke.1211000 and later, you can assign
additional subnets to a VPC-native cluster.

October 07, 2025

      Feature
      Starting with GKE version 1.33.2-gke.1240000 and later, you can specify the
network tier (Standard or Premium) for ephemeral IP addresses.

September 11, 2025

      Feature
      The accelerator-optimized A4X VM is available as a4x-highgpu-4g in the us-central1-a zone with GKE version 1.32.8-gke.1108000 or later.

August 29, 2025

      Security
      A fix is available for an issue with Cloud Storage FUSE CSI driver in GKE versions 1.33.1-gke.1959000 and later, and 1.32.6-gke.1125000 and later.

August 28, 2025

      Security
      GKE version 1.33.0-gke.1276000 and later remediate a low severity vulnerability.
`;

const expectedOutput = `October 28, 2025

      Feature
      You can use the G4 VM, powered by NVIDIA's RTX PRO 6000 GPUs, with
GKE Autopilot in version 1.34.1-gke.1829001 or later.
      Feature
      Autoscaled blue-green upgrades are available in Preview for
GKE Standard node pools.

October 21, 2025

      Feature
      The G4 VM is generally available on GKE.
For GKE Standard, use GKE version
1.34.0-gke.1662000 or later.

October 17, 2025

      Issue
      Don't use GKE version 1.34.1-gke.1431000 or later when creating
or upgrading node pools with the a3-highgpu-8g machine type.

October 14, 2025

      Issue
      In GKE versions 1.32.4-gke.1029000 and later, MountVolume calls
for network file system (NFS) volumes might fail.

October 09, 2025

      Feature
      In GKE version 1.33.4-gke.1055000 or later, you can control
how external traffic reaches your Services on GKE clusters by
using Network Service Tiers.
      Feature
      In GKE version 1.30.3-gke.1211000 and later, you can assign
additional subnets to a VPC-native cluster.

October 07, 2025

      Feature
      Starting with GKE version 1.33.2-gke.1240000 and later, you can specify the
network tier (Standard or Premium) for ephemeral IP addresses.

September 11, 2025

      Feature
      The accelerator-optimized A4X VM is available as a4x-highgpu-4g in the us-central1-a zone with GKE version 1.32.8-gke.1108000 or later.

August 29, 2025

      Security
      A fix is available for an issue with Cloud Storage FUSE CSI driver in GKE versions 1.33.1-gke.1959000 and later, and 1.32.6-gke.1125000 and later.

August 28, 2025

      Security
      GKE version 1.33.0-gke.1276000 and later remediate a low severity vulnerability.`;

describe('extractReleaseNotesRelevantForUpgrade', () => {
  it('should extract relevant notes for standard upgrade path', () => {
    const result = extractReleaseNotesRelevantForUpgrade(fullNotes, "1.30.3-gke.1211000", "1.34.1-gke.1431000");
    expect(result.trim()).toBe(expectedOutput.trim());
  });
});
