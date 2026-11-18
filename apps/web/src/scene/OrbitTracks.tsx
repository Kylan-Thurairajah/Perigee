import { Line } from '@react-three/drei';
import { toScenePosition } from './sceneMath';
import type { EncounterVisual, OrbitPathVisual, SampledOrbitPathVisual, SceneSelection, SceneOrbitPathVisual } from './types';

interface OrbitTracksProps {
  paths: readonly OrbitPathVisual[];
  selected: SceneSelection;
  encounter?: EncounterVisual | null;
  palette: 'light' | 'dark';
}

export function OrbitTracks({ paths, selected, encounter, palette }: OrbitTracksProps) {
  return (
    <>
      {paths.map((path, index) => {
        if (path.points.length < 2) return null;
        const sampled = 'satelliteId' in path;
        const satelliteId = sampled ? path.satelliteId : path.id;
        const isSelectedSatellite = selected?.kind === 'satellite' && selected.id === satelliteId;
        const isEncounterSatellite = selected?.kind === 'encounter'
          && (satelliteId === encounter?.objectAId || satelliteId === encounter?.objectBId);
        const isDimmedEncounterTrack = selected?.kind === 'encounter' && !isEncounterSatellite;
        const sampledPath = path as SampledOrbitPathVisual;
        const fixturePath = path as SceneOrbitPathVisual;
        const variantColor = sampled
          ? sampledPath.variant === 'simulated' ? '#eba16f' : undefined
          : fixturePath.tone === 'hypothetical' ? '#eba16f' : undefined;
        const color = (sampled ? sampledPath.color : undefined) ?? variantColor
          ?? (!sampled && fixturePath.tone === 'selected' ? '#5f88ad' : undefined)
          ?? (palette === 'dark' ? '#86a9c5' : '#829fb9');
        return (
          <Line
            key={`${satelliteId}-${sampled ? sampledPath.variant ?? 'original' : fixturePath.tone}-${index}`}
            points={sampled
              ? sampledPath.points.map((point) => toScenePosition(point.positionKm))
              : fixturePath.points.map((point) => [...point])}
            color={color}
            lineWidth={isSelectedSatellite ? 1.8 : 1}
            transparent
            opacity={isSelectedSatellite || isEncounterSatellite ? 0.92 : isDimmedEncounterTrack ? 0.2 : 0.62}
            dashed={sampled ? sampledPath.variant === 'simulated' : fixturePath.tone === 'hypothetical'}
            dashSize={0.09}
            gapSize={0.055}
          />
        );
      })}
    </>
  );
}
