import { Suspense, useEffect, useMemo, useState } from 'react';
import { Canvas } from '@react-three/fiber';
import * as THREE from 'three';
import { CameraRig } from './CameraRig';
import { Earth } from './Earth';
import { EncounterMarker } from './EncounterMarker';
import { OrbitTracks } from './OrbitTracks';
import { Satellite } from './Satellite';
import { fixturePositionKm, interpolatePathPosition, toScenePosition } from './sceneMath';
import type {
  OrbitPathVisual,
  PositionKm,
  SatelliteVisual,
  SceneProps,
} from './types';

function usePrefersReducedMotion(): boolean {
  const [reducedMotion, setReducedMotion] = useState(false);
  useEffect(() => {
    const preference = window.matchMedia('(prefers-reduced-motion: reduce)');
    const updatePreference = () => setReducedMotion(preference.matches);
    updatePreference();
    preference.addEventListener?.('change', updatePreference);
    return () => preference.removeEventListener?.('change', updatePreference);
  }, []);
  return reducedMotion;
}

function getPathForSatellite(satelliteId: string, paths: readonly OrbitPathVisual[]): OrbitPathVisual | undefined {
  return paths.find((path) => 'satelliteId' in path && path.satelliteId === satelliteId && path.variant !== 'simulated')
    ?? paths.find((path) => ('satelliteId' in path ? path.satelliteId : path.id) === satelliteId);
}

function getCurrentPositionScene(
  satellite: SatelliteVisual,
  paths: readonly OrbitPathVisual[],
  timeUtc: string,
): [number, number, number] {
  const path = getPathForSatellite(satellite.id, paths);
  if (path && 'satelliteId' in path) {
    const positionKm = interpolatePathPosition(path.points, timeUtc) ?? satellite.positionKm;
    if (positionKm) return toScenePosition(positionKm);
  }
  if (path && !('satelliteId' in path) && path.points.length > 0) {
    const point = path.points[0];
    return [point[0], point[1], point[2]];
  }
  const positionKm: PositionKm = satellite.positionKm ?? fixturePositionKm(satellite.id, satellite.altitudeKm);
  return toScenePosition(positionKm);
}

export function OrbitalScene({
  mode,
  sceneTimeUtc,
  selected,
  satellites,
  paths,
  encounter,
  palette = 'light',
  reducedMotion: reducedMotionOverride,
  onSelect,
}: SceneProps) {
  const systemReducedMotion = usePrefersReducedMotion();
  const reducedMotion = reducedMotionOverride ?? systemReducedMotion;
  const [interactionPaused, setInteractionPaused] = useState(false);

  const positionsScene = useMemo(
    () => new Map(satellites.map((satellite) => [
      satellite.id,
      getCurrentPositionScene(satellite, paths, sceneTimeUtc),
    ])),
    [paths, satellites, sceneTimeUtc],
  );
  const focusPositions = useMemo(() => {
    const focuses = new Map(positionsScene);
    if (encounter) {
      const first = positionsScene.get(encounter.objectAId);
      const second = positionsScene.get(encounter.objectBId);
      if (first && second) {
        focuses.set(encounter.id, [
          (first[0] + second[0]) / 2,
          (first[1] + second[1]) / 2,
          (first[2] + second[2]) / 2,
        ]);
      }
    }
    return focuses;
  }, [encounter, positionsScene]);
  const encounterPosition = useMemo(() => {
    if (!encounter) return undefined;
    const firstSatellite = satellites.find((satellite) => satellite.id === encounter.objectAId);
    const secondSatellite = satellites.find((satellite) => satellite.id === encounter.objectBId);
    if (!firstSatellite || !secondSatellite) return undefined;
    const first = getCurrentPositionScene(firstSatellite, paths, encounter.tcaUtc);
    const second = getCurrentPositionScene(secondSatellite, paths, encounter.tcaUtc);
    return new THREE.Vector3(
      (first[0] + second[0]) / 2,
      (first[1] + second[1]) / 2,
      (first[2] + second[2]) / 2,
    ).toArray() as [number, number, number];
  }, [encounter, paths, satellites]);
  const encounterSatelliteIds = useMemo(
    () => new Set(encounter ? [encounter.objectAId, encounter.objectBId] : []),
    [encounter],
  );
  const selectedSatelliteId = selected?.kind === 'satellite' ? selected.id : undefined;
  const selectedEncounterId = selected?.kind === 'encounter' ? selected.id : undefined;

  return (
    <div
      className="perigee-orbital-scene"
      style={{ width: '100%', height: '100%', minHeight: 320 }}
      aria-label="Interactive orbital scene"
      onPointerDown={() => setInteractionPaused(true)}
      onPointerUp={() => setInteractionPaused(false)}
      onPointerLeave={() => setInteractionPaused(false)}
    >
      <Canvas
        camera={{ position: [8, 6, 9], fov: 36, near: 0.1, far: 100 }}
        dpr={[1, 1.75]}
        onPointerMissed={() => onSelect(null)}
        gl={{ antialias: true, alpha: true, powerPreference: 'high-performance' }}
      >
        <color attach="background" args={[palette === 'dark' ? '#151f29' : '#f7f6f2']} />
        <ambientLight intensity={palette === 'dark' ? 1.05 : 1.3} />
        <directionalLight position={[7, 8, 6]} intensity={palette === 'dark' ? 1.3 : 1.7} />
        <directionalLight position={[-5, -2, -4]} intensity={0.45} color="#c7d9e3" />

        <Suspense fallback={null}>
          <Earth palette={palette} reducedMotion={reducedMotion} paused={interactionPaused} />
          <OrbitTracks paths={paths} selected={selected} encounter={encounter} palette={palette} />
          {satellites.map((satellite) => {
            const position = positionsScene.get(satellite.id) ?? toScenePosition(
              satellite.positionKm ?? fixturePositionKm(satellite.id, satellite.altitudeKm),
            );
            const dimmed = selected?.kind === 'encounter'
              ? !encounterSatelliteIds.has(satellite.id)
              : Boolean(selectedSatelliteId && selectedSatelliteId !== satellite.id);
            return (
              <Satellite
                key={satellite.id}
                satellite={satellite}
                position={position}
                selected={selectedSatelliteId === satellite.id}
                dimmed={dimmed}
                palette={palette}
                onSelect={onSelect}
              />
            );
          })}
          {encounter && encounterPosition && (
            <EncounterMarker
              position={encounterPosition}
              encounterId={encounter.id}
              selected={selectedEncounterId === encounter.id}
              onSelect={onSelect}
            />
          )}
        </Suspense>
        <CameraRig
          mode={mode}
          selected={selected}
          focusPositions={focusPositions}
          reducedMotion={reducedMotion}
        />
      </Canvas>
    </div>
  );
}

export default OrbitalScene;
