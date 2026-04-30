import { describe, it, expect } from 'vitest';

describe('ClusterToolkit Args', () => {
  it('should create clusterToolkitDownloadArgs', () => {
    const args = {
      download_directory: "/tmp/cluster-toolkit",
    };

    expect(args.download_directory).toBe("/tmp/cluster-toolkit");
  });

  it('should handle empty download_directory', () => {
    const args = {
      download_directory: "",
    };
    expect(args.download_directory).toBe("");
  });

  it('should handle trailing slash', () => {
    const args = {
      download_directory: "/home/user/",
    };
    expect(args.download_directory).toBe("/home/user/");
  });

  it('should handle custom path', () => {
    const args = {
      download_directory: "/opt/cluster-toolkit-download",
    };
    expect(args.download_directory).toBe("/opt/cluster-toolkit-download");
  });

  it('should handle relative path', () => {
    const args = {
      download_directory: "./downloads",
    };
    expect(args.download_directory).toBe("./downloads");
  });
});
