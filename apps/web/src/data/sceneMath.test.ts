import { describe, expect, it } from 'vitest';
import {
  EARTH_RADIUS_KM,
  EARTH_RADIUS_SCENE,
  ORBIT_VISUAL_CLEARANCE_SCENE,
  fixturePositionKm,
  impactCameraPosition,
  interpolatePathPosition,
  toScenePosition,
} from '../scene/sceneMath';
import type { OrbitPathPoint } from '../scene/types';

const points: OrbitPathPoint[] = [
  { atUtc: '2026-03-20T12:00:00Z', positionKm: [7000, 0, 0] },
  { atUtc: '2026-03-20T13:00:00Z', positionKm: [0, 7000, 0] },
];

describe('sceneMath', () => {
  it('places the impact camera outside the event even on the far side of Earth', () => {
    for (const impact of [[-2.2, 0, 0], [0, 0, -2.3], [0, -2.4, 0]] as const) {
      const camera = impactCameraPosition(impact, 7.6);
      const midpoint = camera.map((coordinate, index) => (coordinate + impact[index]) / 2);
      expect(Math.hypot(...midpoint)).toBeGreaterThan(Math.hypot(...impact));
      expect(Math.hypot(...camera.map((coordinate, index) => coordinate - impact[index]))).toBeCloseTo(7.6);
    }
  });
  it('interpolates synchronized path samples at the requested UTC time', () => {
    expect(interpolatePathPosition(points, '2026-03-20T12:30:00Z')).toEqual([3500, 3500, 0]);
  });

  it('clamps times outside a path and safely falls back for invalid UTC input', () => {
    expect(interpolatePathPosition(points, '2026-03-20T11:00:00Z')).toBe(points[0].positionKm);
    expect(interpolatePathPosition(points, '2026-03-20T14:00:00Z')).toBe(points[1].positionKm);
    expect(interpolatePathPosition(points, 'not-a-time')).toBe(points[0].positionKm);
  });

  it('maps Earth-centered kilometers into the scene scale', () => {
    expect(toScenePosition([EARTH_RADIUS_KM, 0, 0])[0]).toBeCloseTo(EARTH_RADIUS_SCENE);
    expect(toScenePosition([EARTH_RADIUS_KM, 0, 0]).slice(1)).toEqual([0, 0]);
  });

  it('adds visual clearance to orbit paths without changing their direction', () => {
    const position = toScenePosition([EARTH_RADIUS_KM + 550, 0, 0]);
    const expectedRadius = ((EARTH_RADIUS_KM + 550) * EARTH_RADIUS_SCENE / EARTH_RADIUS_KM)
      + ORBIT_VISUAL_CLEARANCE_SCENE;
    expect(position[0]).toBeCloseTo(expectedRadius);
    expect(position.slice(1)).toEqual([0, 0]);
  });

  it('creates a deterministic visual fixture position at the requested altitude', () => {
    const first = fixturePositionKm('sat-peregrine-a', 550);
    expect(fixturePositionKm('sat-peregrine-a', 550)).toEqual(first);
    expect(Math.hypot(...first)).toBeCloseTo(EARTH_RADIUS_KM + 550);
    expect(fixturePositionKm('sat-lyra-b', 550)).not.toEqual(first);
  });
});
