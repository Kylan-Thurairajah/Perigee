export type SceneMode = 'explore' | 'encounter' | 'maneuver';

export type SceneSelection =
  | { kind: 'satellite'; id: string }
  | { kind: 'encounter'; id: string }
  | null;

export type PositionKm = readonly [x: number, y: number, z: number];

export interface SatelliteVisual {
  id: string;
  name: string;
  altitudeKm: number;
  /** Earth-centered position in kilometers, matching the API trajectory frame. */
  positionKm?: PositionKm;
  color?: string;
}

export interface OrbitPathPoint {
  /** RFC3339 UTC time for this synchronized API trajectory sample. */
  atUtc: string;
  positionKm: PositionKm;
}

export interface SampledOrbitPathVisual {
  satelliteId: string;
  points: readonly OrbitPathPoint[];
  color?: string;
  /** Maneuver paths can be styled distinctly from the original orbit. */
  variant?: 'original' | 'simulated';
}

export interface SceneOrbitPathVisual {
  /** Satellite identifier, matching the existing SceneSlot contract. */
  id: string;
  /** Positions already expressed in scene units. */
  points: readonly PositionKm[];
  tone: 'orbit' | 'selected' | 'hypothetical';
}

export type OrbitPathVisual = SampledOrbitPathVisual | SceneOrbitPathVisual;

export interface EncounterVisual {
  id: string;
  tcaUtc: string;
  objectAId: string;
  objectBId: string;
}

export interface SceneProps {
  mode: SceneMode;
  sceneTimeUtc: string;
  selected: SceneSelection;
  satellites: readonly SatelliteVisual[];
  paths: readonly OrbitPathVisual[];
  encounter?: EncounterVisual | null;
  palette?: 'light' | 'dark';
  reducedMotion?: boolean;
  /** Pass the same position samples used by the timeline and separation chart. */
  onSelect: (selection: SceneSelection) => void;
}
