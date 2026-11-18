import { describe, expect, it } from 'vitest';
import { createFixtureOrbitPaths, createFixtureSeparationSamples } from './fixtureOrbitPaths';
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
});
