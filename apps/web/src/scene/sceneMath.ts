import type { OrbitPathPoint, PositionKm } from './types';

export const EARTH_RADIUS_KM = 6371;
export const EARTH_RADIUS_SCENE = 2;
export const KM_TO_SCENE = EARTH_RADIUS_SCENE / EARTH_RADIUS_KM;
export const ORBIT_VISUAL_CLEARANCE_SCENE = 0.13;

export function toScenePosition(positionKm: PositionKm): [number, number, number] {
  const scenePosition: [number, number, number] = [
    positionKm[0] * KM_TO_SCENE,
    positionKm[1] * KM_TO_SCENE,
    positionKm[2] * KM_TO_SCENE,
  ];
  const radius = Math.hypot(...scenePosition);
  if (radius <= EARTH_RADIUS_SCENE) return scenePosition;

  const visualRadius = radius + ORBIT_VISUAL_CLEARANCE_SCENE;
  const scale = visualRadius / radius;
  return [
    scenePosition[0] * scale,
    scenePosition[1] * scale,
    scenePosition[2] * scale,
  ];
}

export function fixturePositionKm(id: string, altitudeKm: number): PositionKm {
  let hash = 0;
  for (let index = 0; index < id.length; index += 1) {
    hash = (hash * 31 + id.charCodeAt(index)) | 0;
  }
  const angle = ((hash >>> 0) / 0xffffffff) * Math.PI * 2;
  const radiusKm = EARTH_RADIUS_KM + Math.max(0, altitudeKm);
  const inclination = ((hash >>> 8) % 31 - 15) * (Math.PI / 180);
  return [
    radiusKm * Math.cos(inclination) * Math.cos(angle),
    radiusKm * Math.sin(inclination),
    radiusKm * Math.cos(inclination) * Math.sin(angle),
  ];
}

export function interpolatePathPosition(
  points: readonly OrbitPathPoint[],
  timeUtc: string,
): PositionKm | undefined {
  if (points.length === 0) return undefined;
  if (points.length === 1) return points[0].positionKm;

  const targetTime = Date.parse(timeUtc);
  if (!Number.isFinite(targetTime)) return points[0].positionKm;

  const firstTime = Date.parse(points[0].atUtc);
  const lastTime = Date.parse(points[points.length - 1].atUtc);
  if (!Number.isFinite(firstTime) || !Number.isFinite(lastTime)) return points[0].positionKm;
  if (targetTime <= firstTime) return points[0].positionKm;
  if (targetTime >= lastTime) return points[points.length - 1].positionKm;

  for (let index = 1; index < points.length; index += 1) {
    const previous = points[index - 1];
    const next = points[index];
    const previousTime = Date.parse(previous.atUtc);
    const nextTime = Date.parse(next.atUtc);
    if (targetTime <= nextTime) {
      const span = nextTime - previousTime;
      if (span <= 0) return next.positionKm;
      const amount = (targetTime - previousTime) / span;
      return [
        previous.positionKm[0] + (next.positionKm[0] - previous.positionKm[0]) * amount,
        previous.positionKm[1] + (next.positionKm[1] - previous.positionKm[1]) * amount,
        previous.positionKm[2] + (next.positionKm[2] - previous.positionKm[2]) * amount,
      ];
    }
  }
  return points[points.length - 1].positionKm;
}
