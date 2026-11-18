import type { CSSProperties } from 'react';

export type SceneMode = 'explore' | 'encounter' | 'maneuver';
export type SceneSelection =
  | { kind: 'satellite'; id: string }
  | { kind: 'encounter'; id: string }
  | null;

export interface SatelliteVisual {
  id: string;
  name: string;
  altitudeKm: number;
}

export interface OrbitPathVisual {
  id: string;
  points: readonly [number, number, number][];
  tone: 'orbit' | 'selected' | 'hypothetical';
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
  onSelect: (selection: SceneSelection) => void;
}

const orbitRings = [
  { className: 'orbit orbit-one', label: 'Peregrine A orbit' },
  { className: 'orbit orbit-two', label: 'Lyra B orbit' },
  { className: 'orbit orbit-three', label: 'Celeste C orbit' },
];

export default function SceneSlot({
  mode,
  sceneTimeUtc,
  selected,
  satellites,
  paths,
  encounter,
  onSelect,
}: SceneSlotProps) {
  const featuredId = encounter?.id;

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
      <div className="space-art" aria-hidden="true">
        <div className="star-field">
          <i className="star star-a" /><i className="star star-b" /><i className="star star-c" />
          <i className="star star-d" /><i className="star star-e" /><i className="star star-f" />
          <i className="star star-g" /><i className="star star-h" />
        </div>
        {orbitRings.map((orbit) => <div className={orbit.className} key={orbit.className} />)}
        <div className="earth-halo" />
        <div className="earth">
          <div className="earth-grid" />
          <div className="land land-one" />
          <div className="land land-two" />
          <div className="land land-three" />
          <div className="cloud-cloud cloud-one" />
          <div className="cloud-cloud cloud-two" />
        </div>
        <button
          className={`satellite satellite-peregrine ${selected?.kind === 'satellite' && selected.id === 'sat-peregrine-a' ? 'is-selected' : ''}`}
          style={{ '--satellite-delay': '0s' } as CSSProperties}
          aria-label="Select Peregrine A satellite"
          onClick={() => onSelect({ kind: 'satellite', id: 'sat-peregrine-a' })}
        >
          <span className="sat-body" /><span className="sat-wing sat-wing-left" /><span className="sat-wing sat-wing-right" />
        </button>
        <button
          className={`satellite satellite-lyra ${selected?.kind === 'satellite' && selected.id === 'sat-lyra-b' ? 'is-selected' : ''}`}
          style={{ '--satellite-delay': '1.3s' } as CSSProperties}
          aria-label="Select Lyra B satellite"
          onClick={() => onSelect({ kind: 'satellite', id: 'sat-lyra-b' })}
        >
          <span className="sat-body" /><span className="sat-wing sat-wing-left" /><span className="sat-wing sat-wing-right" />
        </button>
        <button
          className={`encounter-marker ${selected?.kind === 'encounter' && selected.id === featuredId ? 'is-selected' : ''}`}
          aria-label="Open featured Peregrine A and Lyra B encounter"
          onClick={() => encounter && onSelect({ kind: 'encounter', id: encounter.id })}
        >
          <span className="marker-pulse" /><span className="marker-diamond" />
        </button>
        <div className="scene-caption">
          <span className="caption-line" />
          <span>{satellites.length} OBJECTS IN VIEW</span>
        </div>
        <div className="earth-label">EARTH<br /><span>LEO · 550 KM</span></div>
      </div>
      <div className="scene-footnote">
        <span>Stylized orbital view</span>
        <span>Not to scale</span>
      </div>
    </section>
  );
}
