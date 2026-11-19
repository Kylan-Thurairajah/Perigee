import { EARTH_RADIUS_KM } from '../scene/sceneMath';
import type { OrbitPathPoint, PositionKm, SampledOrbitPathVisual } from '../scene/types';
import type { CatalogObject, Encounter } from './fixtures';

const EARTH_MU_KM3_S2 = 398600.4418;
const SAMPLE_STEP_SECONDS = 120;
const HALF_WINDOW_MINUTES = 30;

function hashId(id: string): number {
  let hash = 0;
  for (let index = 0; index < id.length; index += 1) {
    hash = (hash * 31 + id.charCodeAt(index)) | 0;
  }
  return hash >>> 0;
}

function circularPosition(radiusKm: number, angle: number, inclination: number): PositionKm {
  return [
    radiusKm * Math.cos(angle),
    radiusKm * Math.sin(angle) * Math.cos(inclination),
    radiusKm * Math.sin(angle) * Math.sin(inclination),
  ];
}

function createEncounterPositions(
  objectA: CatalogObject,
  objectB: CatalogObject,
  encounter: Encounter,
  offsetSeconds: number,
): [PositionKm, PositionKm] {
  const missDistanceKm = encounter.minimumSeparationMeters / 1000;
  const meanCatalogRadiusKm = EARTH_RADIUS_KM + (objectA.meanAltitudeKm + objectB.meanAltitudeKm) / 2;
  const radiusA = meanCatalogRadiusKm - missDistanceKm / 2;
  const radiusB = meanCatalogRadiusKm + missDistanceKm / 2;
  const speedA = Math.sqrt(EARTH_MU_KM3_S2 / radiusA);
  const speedB = Math.sqrt(EARTH_MU_KM3_S2 / radiusB);
  const relativeSpeedKmS = encounter.relativeSpeedMps / 1000;
  const planeCosine = Math.max(-1, Math.min(1,
    (speedA ** 2 + speedB ** 2 - relativeSpeedKmS ** 2) / (2 * speedA * speedB),
  ));
  const planeAngle = Math.acos(planeCosine);
  const angularSpeedA = speedA / radiusA;
  const angularSpeedB = speedB / radiusB;
  const angleA = angularSpeedA * offsetSeconds;
  const angleB = angularSpeedB * offsetSeconds;

  return [
    circularPosition(radiusA, angleA, 0),
    [
      radiusB * Math.cos(angleB),
      radiusB * Math.sin(angleB) * Math.cos(planeAngle),
      radiusB * Math.sin(angleB) * Math.sin(planeAngle),
    ],
  ];
}

function createOrbitPosition(object: CatalogObject, offsetSeconds: number): PositionKm {
  const radiusKm = EARTH_RADIUS_KM + object.meanAltitudeKm;
  const speedKmS = Math.sqrt(EARTH_MU_KM3_S2 / radiusKm);
  const seed = hashId(object.catalogId);
  const phase = (seed / 0xffffffff) * Math.PI * 2;
  const inclination = ((seed >>> 8) % 31) * (Math.PI / 180);
  return circularPosition(radiusKm, phase + (speedKmS / radiusKm) * offsetSeconds, inclination);
}

export function createFixtureOrbitPaths(
  objects: readonly CatalogObject[],
  encounter: Encounter,
): SampledOrbitPathVisual[] {
  const tcaMs = Date.parse(encounter.tcaUtc);
  if (!Number.isFinite(tcaMs)) throw new Error(`Invalid fixture TCA: ${encounter.tcaUtc}`);

  const objectA = objects.find(({ catalogId }) => catalogId === encounter.objectAId);
  const objectB = objects.find(({ catalogId }) => catalogId === encounter.objectBId);
  if (!objectA || !objectB) {
    throw new Error(`Fixture encounter ${encounter.encounterId} references an unknown catalog object`);
  }

  const sampleCount = (HALF_WINDOW_MINUTES * 60 * 2) / SAMPLE_STEP_SECONDS + 1;
  return objects.map((object) => ({
    satelliteId: object.catalogId,
    variant: 'original',
    points: Array.from({ length: sampleCount }, (_, index): OrbitPathPoint => {
      const offsetSeconds = (index * SAMPLE_STEP_SECONDS) - HALF_WINDOW_MINUTES * 60;
      const positionKm = object.catalogId === objectA.catalogId || object.catalogId === objectB.catalogId
        ? createEncounterPositions(objectA, objectB, encounter, offsetSeconds)[
          object.catalogId === objectA.catalogId ? 0 : 1
        ]
        : createOrbitPosition(object, offsetSeconds);
      return {
        atUtc: new Date(tcaMs + offsetSeconds * 1000).toISOString(),
        positionKm,
      };
    }),
  }));
}

export function createAlongTrackPreviewPath(
  path: SampledOrbitPathVisual,
  deltaVMetersPerSecond: number,
  impulseAtUtc: string,
): SampledOrbitPathVisual {
  if (!Number.isFinite(deltaVMetersPerSecond)) throw new Error('Fixture delta-v must be finite');
  if (path.points.length < 2) throw new Error('Fixture maneuver path requires at least two samples');

  const impulseTime = Date.parse(impulseAtUtc);
  const firstTime = Date.parse(path.points[0].atUtc);
  const lastTime = Date.parse(path.points[path.points.length - 1].atUtc);
  if (!Number.isFinite(impulseTime) || impulseTime < firstTime || impulseTime > lastTime) {
    throw new Error(`Fixture impulse time is outside the path: ${impulseAtUtc}`);
  }

  const points = path.points.map((point, index): OrbitPathPoint => {
    const pointTime = Date.parse(point.atUtc);
    if (!Number.isFinite(pointTime)) throw new Error(`Invalid fixture path time: ${point.atUtc}`);

    const previous = path.points[Math.max(0, index - 1)];
    const next = path.points[Math.min(path.points.length - 1, index + 1)];
    const tangent = [
      next.positionKm[0] - previous.positionKm[0],
      next.positionKm[1] - previous.positionKm[1],
      next.positionKm[2] - previous.positionKm[2],
    ] as const;
    const tangentLength = Math.hypot(...tangent);
    if (!Number.isFinite(tangentLength) || tangentLength === 0) {
      throw new Error(`Fixture path has no along-track direction at sample ${index}`);
    }

    const offsetKm = deltaVMetersPerSecond * Math.max(0, (pointTime - impulseTime) / 1000) / 1000;
    return {
      atUtc: point.atUtc,
      positionKm: [
        point.positionKm[0] + (tangent[0] / tangentLength) * offsetKm,
        point.positionKm[1] + (tangent[1] / tangentLength) * offsetKm,
        point.positionKm[2] + (tangent[2] / tangentLength) * offsetKm,
      ],
    };
  });

  return { ...path, variant: 'simulated', points };
}

export function createFixtureSeparationSamples(
  paths: readonly SampledOrbitPathVisual[],
  objectAId: string,
  objectBId: string,
  objectBVariant: 'original' | 'simulated' = 'original',
): number[] {
  const pathA = paths.find(({ satelliteId, variant }) => satelliteId === objectAId && (variant ?? 'original') === 'original');
  const pathB = paths.find(({ satelliteId, variant }) => satelliteId === objectBId && (variant ?? 'original') === objectBVariant);
  if (!pathA || !pathB) throw new Error('Fixture encounter paths are missing an object');
  if (pathA.points.length !== pathB.points.length) throw new Error('Fixture encounter paths are not aligned');

  return pathA.points.map((pointA, index) => {
    const pointB = pathB.points[index];
    if (pointA.atUtc !== pointB.atUtc) throw new Error('Fixture encounter path timestamps are not aligned');
    const dx = pointA.positionKm[0] - pointB.positionKm[0];
    const dy = pointA.positionKm[1] - pointB.positionKm[1];
    const dz = pointA.positionKm[2] - pointB.positionKm[2];
    return Math.sqrt(dx ** 2 + dy ** 2 + dz ** 2) * 1000;
  });
}
