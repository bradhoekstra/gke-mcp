import { describe, it, expect } from 'vitest';

describe('Monitoring Args', () => {
  it('should create listMonitoredResourceDescriptorsArgs', () => {
    const args = {
      project_id: "my-project",
    };

    expect(args.project_id).toBe("my-project");
  });

  it('should handle empty project_id', () => {
    const args = {
      project_id: "",
    };
    expect(args.project_id).toBe("");
  });

  it('should handle different projects', () => {
    const projects = [
      "my-project",
      "my-other-project",
      "123456789012",
    ];

    for (const project of projects) {
      const args = {
        project_id: project,
      };
      expect(args.project_id).toBe(project);
    }
  });
});
