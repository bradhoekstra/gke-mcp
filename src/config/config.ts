import { execSync } from 'child_process';

export class Config {
  private _userAgent: string;
  private _defaultProjectID: string;
  private _defaultLocation: string;

  constructor(version: string) {
    this._userAgent = `gke-mcp/${version}`;
    this._defaultProjectID = this.getDefaultProjectID();
    this._defaultLocation = this.getDefaultLocation();
  }

  get userAgent(): string {
    return this._userAgent;
  }

  get defaultProjectID(): string {
    return this._defaultProjectID;
  }

  get defaultLocation(): string {
    return this._defaultLocation;
  }

  private getDefaultProjectID(): string {
    try {
      return this.getGcloudConfig('core/project');
    } catch (error) {
      console.error(`Failed to get default project: ${error}`);
      return '';
    }
  }

  private getDefaultLocation(): string {
    try {
      const region = this.getGcloudConfig('compute/region');
      if (region) return region;
    } catch (error) {
      // Ignore error and try zone
    }

    try {
      const zone = this.getGcloudConfig('compute/zone');
      if (zone) return zone;
    } catch (error) {
      // Ignore error
    }

    return '';
  }

  private getGcloudConfig(key: string): string {
    try {
      const out = execSync(`gcloud config get ${key}`, { encoding: 'utf8' });
      return out.trim();
    } catch (error) {
      throw error;
    }
  }
}
