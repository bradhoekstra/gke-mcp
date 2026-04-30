import { describe, it, expect } from 'vitest';
import { mapTimeseriesDataPoints, getChartsTools } from './charts.js';
import { Config } from '../config/config.js';

describe('Charts App', () => {
  describe('mapTimeseriesDataPoints', () => {
    it('should handle empty input', () => {
      const result = mapTimeseriesDataPoints({});
      expect(result.label).toBe('');
      expect(result.points).toEqual([]);
    });

    it('should map single point with value and label', () => {
      const now = new Date();
      const input = {
        labelValues: [{ stringValue: "label-1" }],
        pointData: [
          {
            values: [{ doubleValue: 42.5 }],
            timeInterval: { endTime: now.toISOString() },
          },
        ],
      };

      const result = mapTimeseriesDataPoints(input);
      expect(result.label).toBe("label-1");
      expect(result.points.length).toBe(1);
      expect(result.points[0].value).toBe(42.5);
      expect(result.points[0].timestamp).toBe(now.getTime());
    });

    it('should map multiple points and labels', () => {
      const now = new Date();
      const input = {
        labelValues: [
          { stringValue: "foo" },
          { stringValue: "bar" },
        ],
        pointData: [
          {
            values: [{ doubleValue: 1.0 }],
            timeInterval: { endTime: now.toISOString() },
          },
          {
            values: [{ doubleValue: 2.0 }],
            timeInterval: { endTime: new Date(now.getTime() + 3600000).toISOString() },
          },
        ],
      };

      const result = mapTimeseriesDataPoints(input);
      expect(result.label).toBe("foo bar");
      expect(result.points.length).toBe(2);
      expect(result.points[0].value).toBe(1.0);
      expect(result.points[1].value).toBe(2.0);
    });

    it('should skip points with no values', () => {
      const input = {
        pointData: [
          {
            values: [],
            timeInterval: { endTime: new Date().toISOString() },
          },
        ],
      };

      const result = mapTimeseriesDataPoints(input);
      expect(result.points.length).toBe(0);
    });

    it('should convert int64 value to float', () => {
      const now = new Date();
      const input = {
        pointData: [
          {
            values: [{ int64Value: "42" }],
            timeInterval: { endTime: now.toISOString() },
          },
        ],
      };

      const result = mapTimeseriesDataPoints(input);
      expect(result.points[0].value).toBe(42.0);
    });
  });

  describe('Validation', () => {
    const config = new Config('1.0.0');
    const tools = getChartsTools(config);
    const chartTool = tools.find(t => t.name === 'monitoring_time_series_chart');
    const queryTool = tools.find(t => t.name === 'query_time_series');

    if (!chartTool || !queryTool) {
      throw new Error('Tools not found');
    }

    it('should throw error for empty query in chart tool', async () => {
      await expect(chartTool.handler({ project_id: 'test-proj', query: '' }))
        .rejects.toThrow('query argument cannot be empty');
    });

    it('should throw error for empty query in query tool', async () => {
      await expect(queryTool.handler({ project_id: 'test-proj', query: '' }))
        .rejects.toThrow('query argument cannot be empty');
    });
  });
});
