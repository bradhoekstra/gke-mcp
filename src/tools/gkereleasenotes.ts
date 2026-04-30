import { Config } from '../config/config.js';
import * as fs from 'fs';
import * as path from 'path';
import * as cheerio from 'cheerio';

export interface ToolDefinition {
  name: string;
  description: string;
  inputSchema: any;
  handler: (args: any) => Promise<any>;
}

const gkeVersionRegexp = /\d+\.\d+\.\d+-gke\.\d+/g;
const releaseDateHeadingRegexp = /(^|\n)\s*[A-Za-z]+\s+\d+,\s+\d+\s*(\n|$)/g;

function parseGkeVersion(version: string): [number, number, number, number] {
  const parts = version.split('-gke.');
  if (parts.length !== 2) {
    throw new Error(`invalid GKE version format: ${version}`);
  }

  const k8sParts = parts[0].split('.');
  if (k8sParts.length !== 3) {
    throw new Error(`invalid Kubernetes version part in GKE version: ${parts[0]}`);
  }

  const major = parseInt(k8sParts[0], 10);
  const minor = parseInt(k8sParts[1], 10);
  const patch = parseInt(k8sParts[2], 10);
  const gkePatch = parseInt(parts[1], 10);

  if (isNaN(major) || isNaN(minor) || isNaN(patch) || isNaN(gkePatch)) {
    throw new Error(`cannot parse version parts as integers: ${version}`);
  }

  return [major, minor, patch, gkePatch];
}

function compareVersions(a: string, b: string): number {
  const [aMajor, aMinor, aPatch, aGKE] = parseGkeVersion(a);
  const [bMajor, bMinor, bPatch, bGKE] = parseGkeVersion(b);

  if (bMajor !== aMajor) return bMajor > aMajor ? 1 : -1;
  if (bMinor !== aMinor) return bMinor > aMinor ? 1 : -1;
  if (bPatch !== aPatch) return bPatch > aPatch ? 1 : -1;
  if (bGKE !== aGKE) return bGKE > aGKE ? 1 : -1;

  return 0;
}

export function extractReleaseNotesRelevantForUpgrade(fullReleaseNotes: string, sourceVersion: string, targetVersion: string): string {
  const versionMatches = Array.from(fullReleaseNotes.matchAll(gkeVersionRegexp));
  if (!versionMatches.length) return fullReleaseNotes;

  let leftBorderVersionLocation: any = null;
  let rightBorderVersionLocation: any = null;

  for (let i = 0; i < versionMatches.length; i++) {
    const match = versionMatches[i];
    const version = match[0];
    try {
      const cmp = compareVersions(version, targetVersion);
      if (cmp === 0) {
        leftBorderVersionLocation = match;
        break;
      } else if (cmp > 0) {
        if (i === 0) {
          leftBorderVersionLocation = match;
        } else {
          leftBorderVersionLocation = versionMatches[i - 1];
        }
        break;
      }
    } catch (e) {
      continue;
    }
  }

  for (let i = versionMatches.length - 1; i >= 0; i--) {
    const match = versionMatches[i];
    const version = match[0];
    try {
      const cmp = compareVersions(version, sourceVersion);
      if (cmp === 0) {
        rightBorderVersionLocation = match;
        break;
      } else if (cmp < 0) {
        if (i === versionMatches.length - 1) {
          rightBorderVersionLocation = match;
        } else {
          rightBorderVersionLocation = versionMatches[i + 1];
        }
        break;
      }
    } catch (e) {
      continue;
    }
  }

  const leftBorder = leftBorderVersionLocation?.index || 0;
  const rightBorder = rightBorderVersionLocation ? (rightBorderVersionLocation.index! + rightBorderVersionLocation[0].length) : fullReleaseNotes.length;

  let reducedReleaseNotes = fullReleaseNotes.substring(leftBorder, rightBorder);

  const leftCut = fullReleaseNotes.substring(0, leftBorder);
  let leftAppend = '';
  if (leftCut.length > 0) {
    const dateMatches = Array.from(leftCut.matchAll(releaseDateHeadingRegexp));
    if (!dateMatches.length) {
      leftAppend = leftCut;
    } else {
      const lastMatch = dateMatches[dateMatches.length - 1];
      leftAppend = leftCut.substring(lastMatch.index!);
    }
  }

  const rightCut = fullReleaseNotes.substring(rightBorder);
  let rightAppend = '';
  if (rightCut.length > 0) {
    const dateMatches = Array.from(rightCut.matchAll(releaseDateHeadingRegexp));
    if (!dateMatches.length) {
      rightAppend = rightCut;
    } else {
      const firstMatch = dateMatches[0];
      rightAppend = rightCut.substring(0, firstMatch.index!);
    }
  }

  return leftAppend + reducedReleaseNotes + rightAppend;
}

export function getGkeReleaseNotesTools(config: Config): ToolDefinition[] {
  return [
    {
      name: 'get_gke_release_notes',
      description: 'Get GKE release notes. Prefer to use this tool if GKE release notes are needed.',
      inputSchema: {
        type: 'object',
        properties: {
          SourceVersion: { type: 'string', description: "A source GKE version an upgrade happens from. For example, '1.33.5-gke.120000'." },
          TargetVersion: { type: 'string', description: "A target GKE version an upgrade happens from. For example, '1.34.3-gke.240500'." },
        },
        required: ['SourceVersion', 'TargetVersion'],
      },
      handler: async (args: any) => {
        const today = new Date().toISOString().slice(0, 10);
        const releaseNotesFilePath = `release-notes-${today}.html`;
        
        let htmlContent = '';
        
        if (fs.existsSync(releaseNotesFilePath)) {
          console.error(`Reading release notes from cached file: ${releaseNotesFilePath}`);
          htmlContent = fs.readFileSync(releaseNotesFilePath, 'utf8');
        } else {
          console.error('Fetching release notes from web');
          const url = 'https://cloud.google.com/kubernetes-engine/docs/release-notes';
          const response = await fetch(url);
          if (!response.ok) {
            throw new Error(`Failed to get release notes: ${response.statusText}`);
          }
          htmlContent = await response.text();
          fs.writeFileSync(releaseNotesFilePath, htmlContent, { mode: 0o600 });
        }

        const $ = cheerio.load(htmlContent);
        
        $('[data-text$="Version updates"]').parent().parent().remove();
        $('[data-text$="Security updates"]').parent().parent().remove();
        
        let fullText = '';
        $('.releases').each((_, el) => {
          fullText += $(el).text();
        });

        const reducedNotes = extractReleaseNotesRelevantForUpgrade(fullText, args.SourceVersion, args.TargetVersion);

        return {
          content: [
            { type: 'text', text: reducedNotes },
          ],
        };
      },
    },
  ];
}
