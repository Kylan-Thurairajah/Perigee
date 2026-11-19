import { describe, expect, it } from 'vitest';
import {
  createAlongTrackPreviewPath,
  createFixtureOrbitPaths,
  createFixtureSeparationSamples,
} from './fixtureOrbitPaths';
import { catalogObjects, featuredEncounter } from './fixtures';

describe('synthetic fixture replay paths', () => {
  it('aligns the featured pair minimum and sample window with the TCA', () => {
    const paths = createFixtureOrbitPaths(catalogObjects, featuredEncounter);
    const separations = createFixtureSeparationSamples(
      paths,
      featuredEncounter.objectAId,
      featuredEncounter.objectBId,
    );
    const centerIndex = Math.floor(separations.length / 2);

    expect(paths).toHaveLength(catalogObjects.length);
    expect(paths[0].points[0].atUtc).toBe('2026-03-20T14:02:10.000Z');
    expect(paths[0].points[centerIndex].atUtc).toBe('2026-03-20T14:32:10.000Z');
    expect(paths[0].points[paths[0].points.length - 1].atUtc).toBe('2026-03-20T15:02:10.000Z');
    expect(separations[centerIndex]).toBeCloseTo(featuredEncounter.minimumSeparationMeters, 6);
    expect(separations[centerIndex]).toBeLessThan(separations[centerIndex - 1]);
    expect(separations[centerIndex]).toBeLessThan(separations[centerIndex + 1]);
  });

  it('creates a time-aligned linearized preview and derives its TCA separation from the changed path', () => {
    const paths = createFixtureOrbitPaths(catalogObjects, featuredEncounter);
    const targetPath = paths.find(({ satelliteId }) => satelliteId === featuredEncounter.objectBId);
    if (!targetPath) throw new Error('Test fixture target path is missing');

    const previewPath = createAlongTrackPreviewPath(targetPath, 0.35, targetPath.points[0].atUtc);
    const previewSeparations = createFixtureSeparationSamples(
      [...paths, previewPath],
      featuredEncounter.objectAId,
      featuredEncounter.objectBId,
      'simulated',
    );
    const centerIndex = Math.floor(previewSeparations.length / 2);

    expect(previewPath.variant).toBe('simulated');
    expect(previewPath.points[0].positionKm).toEqual(targetPath.points[0].positionKm);
    expect(previewPath.points[centerIndex].positionKm).not.toEqual(targetPath.points[centerIndex].positionKm);
    expect(previewSeparations[centerIndex]).toBeCloseTo(
      Math.hypot(featuredEncounter.minimumSeparationMeters, 0.35 * 30 * 60),
      6,
    );
  });
});
