import { describe, it, expect } from 'vitest';

describe('Cluster Args', () => {
  it('should create listClustersArgs', () => {
    const args = {
      project_id: "test-project",
      location: "us-central1",
    };

    expect(args.project_id).toBe("test-project");
    expect(args.location).toBe("us-central1");
  });

  it('should create getClustersArgs', () => {
    const args = {
      project_id: "test-project",
      location: "us-central1",
      name: "my-cluster",
    };

    expect(args.project_id).toBe("test-project");
    expect(args.location).toBe("us-central1");
    expect(args.name).toBe("my-cluster");
  });

  it('should create createClustersArgs', () => {
    const args = {
      project_id: "test-project",
      location: "us-central1",
      cluster: {
        name: "my-cluster",
      },
    };

    expect(args.project_id).toBe("test-project");
    expect(args.location).toBe("us-central1");
    expect(args.cluster.name).toBe("my-cluster");
  });

  it('should create getKubeconfigArgs', () => {
    const args = {
      project_id: "test-project",
      location: "us-west1",
      name: "my-cluster",
    };

    expect(args.project_id).toBe("test-project");
    expect(args.location).toBe("us-west1");
    expect(args.name).toBe("my-cluster");
  });

  it('should create getNodeSosReportArgs', () => {
    const args = {
      node: "my-node",
      destination: "/tmp/sos",
      method: "pod",
      timeout: 300,
    };

    expect(args.node).toBe("my-node");
    expect(args.destination).toBe("/tmp/sos");
    expect(args.method).toBe("pod");
    expect(args.timeout).toBe(300);
  });
});
