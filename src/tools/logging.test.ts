import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getLoggingTools } from './logging.js';
import { Config } from '../config/config.js';
import * as fs from 'fs';

const mockGetEntries = vi.fn();

vi.mock('@google-cloud/logging', () => {
  return {
    Logging: class {
      getEntries = mockGetEntries;
    },
  };
});

describe('Logging Tools', () => {
  const config = new Config('1.0.0');
  const tools = getLoggingTools(config);
  const queryTool = tools.find(t => t.name === 'query_logs');
  const schemaTool = tools.find(t => t.name === 'get_log_schema');

  if (!queryTool || !schemaTool) {
    throw new Error('Logging tools not found');
  }

  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe('query_logs', () => {
    it('should query logs with basic arguments', async () => {
      mockGetEntries.mockResolvedValue([[]]); // return empty list

      await queryTool.handler({ project_id: 'test-project', query: 'severity=ERROR' });

      expect(mockGetEntries).toHaveBeenCalledWith(expect.objectContaining({
        filter: 'severity=ERROR',
      }));
    });

    it('should handle since parameter', async () => {
      mockGetEntries.mockResolvedValue([[]]);

      await queryTool.handler({ project_id: 'test-project', since: '10s' });

      expect(mockGetEntries).toHaveBeenCalledWith(expect.objectContaining({
        filter: expect.stringContaining('timestamp >='),
      }));
    });

    it('should handle time_range parameter', async () => {
      mockGetEntries.mockResolvedValue([[]]);

      await queryTool.handler({
        project_id: 'test-project',
        time_range: { start_time: '2023-01-01T00:00:00Z', end_time: '2023-01-02T00:00:00Z' }
      });

      expect(mockGetEntries).toHaveBeenCalledWith(expect.objectContaining({
        filter: expect.stringContaining('timestamp >= "2023-01-01T00:00:00Z"'),
      }));
    });

    it('should throw error for limit too high', async () => {
      await expect(queryTool.handler({ project_id: 'test-project', limit: 101 }))
        .rejects.toThrow('limit parameter cannot be greater than 100');
    });
  });

  describe('get_log_schema', () => {
    it('should return schema for valid log type', async () => {
      const result = await schemaTool.handler({ log_type: 'k8s_audit_logs' });
      expect(result.content[0].text).toContain('# Kubernetes Audit Logs Schema');
    });

    it('should throw error for invalid log type', async () => {
      await expect(schemaTool.handler({ log_type: 'invalid_log_type' }))
        .rejects.toThrow('unsupported log_type: invalid_log_type');
    });
  });
});
