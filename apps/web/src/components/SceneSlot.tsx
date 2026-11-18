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
  sceneTimeUtc: string;
  selected: SceneSelection;
  satellites: SatelliteVisual[];
  paths: OrbitPathVisual[];
  encounter?: EncounterVisual;
  palette?: 'light' | 'dark';
  onSelect: (selection: SceneSelection) => void;
}

export default function SceneSlot({
  mode,
  sceneTimeUtc,
  selected,
  satellites,
  paths,
  encounter,
  palette = 'light',
  onSelect,
}: SceneSlotProps) {
  return (
    <section
      className={`scene-slot scene-${mode}`}
      aria-label="Orbital scene preview"
      data-scene-time={sceneTimeUtc}
      data-path-count={paths.length}
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
          <span>{satellites.length} OBJECTS IN VIEW</span>
        </div>
        <div className="earth-label">EARTH<br /><span>LEO · 550 KM</span></div>
      </div>
      <div className="scene-footnote">
        <span>Illustrative fixture tracks</span>
        <span>Not to scale</span>
      </div>
    </section>
  );
}
