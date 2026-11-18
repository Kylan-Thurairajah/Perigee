import catalog from '../../../../data/fixtures/catalog.json';
import encounters from '../../../../data/fixtures/encounters.json';

export type DatasetType = 'synthetic' | 'public_snapshot';

export interface SourceMetadata {
  sourceName: string;
  sourceUrl: string | null;
  retrievedAtUtc: string;
  datasetType: DatasetType;
  snapshotId: string;
  license?: string;
  caveat?: string;
}

export interface CatalogObject {
  catalogId: string;
  name: string;
  catalogSnapshotId: string;
  noradId: number | null;
  sourceEpochUtc: string;
  datasetType: DatasetType;
  orbitRegime: 'leo' | 'meo' | 'geo' | 'other';
  meanAltitudeKm: number;
  dataQuality: 'synthetic_fixture' | 'public_snapshot' | 'stale';
  elementModel: string;
  source: SourceMetadata;
}

export interface Encounter {
  encounterId: string;
  objectAId: string;
  objectBId: string;
  catalogSnapshotId: string;
  runId: string | null;
  tcaUtc: string;
  minimumSeparationMeters: number;
  relativeSpeedMps: number;
  sourceEpochUtc: string;
  datasetType: DatasetType;
  dataQuality: string;
  featured: boolean;
}

interface CatalogFixture {
  catalogSnapshotId: string;
  datasetType: DatasetType;
  snapshot: SourceMetadata;
  generatedSeed: string;
  sourceEpochUtc: string;
  objects: CatalogObject[];
}

interface EncounterFixture {
  catalogSnapshotId: string;
  datasetType: DatasetType;
  featuredEncounterId: string;
  sourceEpochUtc: string;
  source: SourceMetadata;
  generatedSeed: string;
  encounters: Encounter[];
}

export const catalogFixture = catalog as CatalogFixture;
export const encounterFixture = encounters as EncounterFixture;

export const catalogObjects = catalogFixture.objects;
export const encounterList = encounterFixture.encounters;
export const featuredEncounter =
  encounterList.find(({ encounterId }) => encounterId === encounterFixture.featuredEncounterId) ??
  encounterList[0];

export function getObject(catalogId: string): CatalogObject | undefined {
  return catalogObjects.find(({ catalogId: id }) => id === catalogId);
}
