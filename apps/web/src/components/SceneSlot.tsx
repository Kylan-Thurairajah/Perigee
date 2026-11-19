import OrbitalScene from '../scene/OrbitalScene';
import type { OrbitPathVisual, SceneMode, SceneSelection } from '../scene/types';

export type { SceneMode, SceneSelection } from '../scene/types';

export interface SatelliteVisual {
  id: string;
  name: string;
  altitudeKm: number;
}

export interface EncounterVisual {
  id: string;
  tcaUtc: string;
  objectAId: string;
  objectBId: string;
}

export interface SceneSlotProps {
  mode: SceneMode;
  collisionProgress?: number;
  sceneTimeUtc: string;
  selected: SceneSelection;
  satellites: SatelliteVisual[];
  paths: OrbitPathVisual[];
  encounter?: EncounterVisual;
  computed?: boolean;
  palette?: 'light' | 'dark';
  onSelect: (selection: SceneSelection) => void;
}

export default function SceneSlot({
  mode,
  collisionProgress,
  sceneTimeUtc,
  selected,
  satellites,
  paths,
  encounter,
  palette = 'light',
  computed = false,
  onSelect,
}: SceneSlotProps) {
  const hasSimulatedPath = paths.some((path) => 'satelliteId' in path && path.variant === 'simulated');
  const collisionPhase = collisionProgress === undefined ? undefined : collisionProgress < 0 ? 'approach' : collisionProgress < 1 ? 'impact' : 'aftermath';
  const visibleCount = collisionPhase === 'impact' || collisionPhase === 'aftermath'
    ? satellites.filter(({ id }) => id !== encounter?.objectAId && id !== encounter?.objectBId).length
    : satellites.length;

  return (
    <section
      className={`scene-slot scene-${mode}`}
      aria-label="Orbital scene preview"
      data-scene-time={sceneTimeUtc}
      data-path-count={paths.length}
      data-collision-phase={collisionPhase}
    >
      <div className="scene-topline">
        <span className="scene-kicker">
          <span className="live-dot" /> ORBIT VIEW
        </span>
        <span className="scene-index">01 — 04</span>
      </div>
      <div className="space-art">
        <OrbitalScene
          mode={mode}
          collisionProgress={collisionProgress}
          sceneTimeUtc={sceneTimeUtc}
          selected={selected}
          satellites={satellites}
          paths={paths}
          encounter={encounter}
          palette={palette}
          onSelect={onSelect}
        />
        <div className="scene-caption">
          <span className="caption-line" />
          <span>{visibleCount} OBJECTS IN VIEW</span>
        </div>
        <div className="earth-label">EARTH<br /><span>{computed ? 'TEME' : 'LEO · 550 KM'}</span></div>
      </div>
      <div className="scene-footnote">
        <span>{computed ? (hasSimulatedPath ? 'Dashed: two-body what-if' : 'SGP4 replay tracks') : (hasSimulatedPath ? 'Dashed: linearized what-if' : 'Illustrative fixture tracks')}</span>
        {collisionPhase
          ? <span role="status">{collisionPhase === 'approach' ? 'Approaching impact' : collisionPhase === 'impact' ? 'Scripted impact' : 'After impact'}</span>
          : <span>Not to scale</span>}
      </div>
    </section>
  );
}
