import { useEffect, useMemo, useRef, useState, type FormEvent } from 'react';
import SceneSlot, { type SceneSelection } from '../components/SceneSlot';
import SeparationChart from '../components/SeparationChart';
import TimelineScrubber from '../components/TimelineScrubber';
import { EARTH_RADIUS_KM } from '../scene/sceneMath';
import type { SampledOrbitPathVisual } from '../scene/types';
import type { ScreeningEncounter } from './ScreeningPanel';
import { requestJSON } from './screeningApi';

type Position = { positionXKm: number; positionYKm: number; positionZKm: number };

export type ScreeningEncounterDetail = {
  encounter: ScreeningEncounter & { sourceEpochUtc: string };
  objectAName: string;
  objectBName: string;
  trajectory: {
    frame: { frame: 'TEME'; frameCaveat: string };
    startUtc: string;
    endUtc: string;
    samples: { atUtc: string; objectA: Position; objectB: Position; separationMeters: number }[];
    source: { sourceName: string; snapshotId: string; retrievedAtUtc: string; caveat: string };
  };
};

type WhatIfResult = {
  alongTrackDeltaVMps: number;
  impulseAtUtc: string;
  samples: ScreeningEncounterDetail['trajectory']['samples'];
  caveats: string[];
};

export default function ScreeningReplay({ detail, palette, runId }: {
  detail: ScreeningEncounterDetail;
  palette: 'light' | 'dark';
  runId: string;
}) {
  const { encounter, trajectory, objectAName, objectBName } = detail;
  const { samples, source } = trajectory;
  const intervals = samples.length - 1;
  const tcaSample = samples[intervals / 2];
  const [index, setIndex] = useState(intervals / 2);
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);
  const [selected, setSelected] = useState<SceneSelection>({ kind: 'encounter', id: encounter.encounterId });
  const [deltaV, setDeltaV] = useState('0.10');
  const [scenario, setScenario] = useState<WhatIfResult | null>(null);
  const [scenarioError, setScenarioError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const request = useRef<AbortController | null>(null);
  useEffect(() => () => request.current?.abort(), []);
  const firstName = objectAName || encounter.objectAId;
  const secondName = objectBName || encounter.objectBId;
  const paths = useMemo(() => ([
    { key: 'objectA', id: encounter.objectAId },
    { key: 'objectB', id: encounter.objectBId },
  ] as const).map(({ key, id }): SampledOrbitPathVisual => ({
    satelliteId: id,
    points: samples.map(({ atUtc, [key]: position }) => ({
      atUtc,
      positionKm: [position.positionXKm, position.positionYKm, position.positionZKm],
    })),
  })), [samples, encounter.objectAId, encounter.objectBId]);
  const satellites = useMemo(() => [
    { id: encounter.objectAId, name: firstName, position: tcaSample.objectA },
    { id: encounter.objectBId, name: secondName, position: tcaSample.objectB },
  ].map(({ id, name, position }) => ({
    id, name,
    altitudeKm: Math.hypot(position.positionXKm, position.positionYKm, position.positionZKm) - EARTH_RADIUS_KM,
  })), [encounter.objectAId, encounter.objectBId, firstName, secondName, tcaSample]);
  const sample = samples[index];
  const comparison = scenario?.samples[index];
  const baselineSeparation = useMemo(() => samples.map(({ separationMeters }) => separationMeters), [samples]);
  const comparisonSeparation = useMemo(() => scenario?.samples.map(({ separationMeters }) => separationMeters), [scenario]);
  const comparisonPath = useMemo<SampledOrbitPathVisual | null>(() => scenario ? {
    satelliteId: encounter.objectAId,
    variant: 'simulated',
    points: scenario.samples.map(({ atUtc, objectA }) => ({
      atUtc,
      positionKm: [objectA.positionXKm, objectA.positionYKm, objectA.positionZKm],
    })),
  } : null, [scenario, encounter.objectAId]);
  const halfWindowMinutes = (Date.parse(trajectory.endUtc) - Date.parse(trajectory.startUtc)) / 120_000;
  const utcLabel = `${sample.atUtc.slice(11, 19)} UTC`;

  async function preview(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (request.current) return;
    const controller = new AbortController();
    request.current = controller;
    setSubmitting(true);
    setScenarioError('');
    try {
      const result = await requestJSON<WhatIfResult>(
        `/api/screenings/${encodeURIComponent(runId)}/encounters/${encodeURIComponent(encounter.encounterId)}/what-if`,
        {
          method: 'POST', headers: { 'Content-Type': 'application/json' }, signal: controller.signal,
          body: JSON.stringify({ alongTrackDeltaVMps: Number(deltaV) }),
        },
      );
      if (!controller.signal.aborted) setScenario(result);
    } catch (failure) {
      if (!controller.signal.aborted) setScenarioError(failure instanceof Error ? failure.message : 'Could not run the what-if.');
    } finally {
      request.current = null;
      if (!controller.signal.aborted) setSubmitting(false);
    }
  }

  return (
    <section className="screening-results" aria-label="Screened encounter replay">
      <h3>{firstName} × {secondName}</h3>
      <p>Reported TCA <time dateTime={encounter.tcaUtc}>{encounter.tcaUtc}</time> · Miss distance {encounter.minimumSeparationMeters.toFixed(1)} m · Relative speed {encounter.relativeSpeedMps.toFixed(1)} m/s</p>
      <SceneSlot
        computed
        mode="encounter"
        sceneTimeUtc={sample.atUtc}
        selected={selected}
        satellites={satellites}
        paths={comparisonPath ? [...paths, comparisonPath] : paths}
        encounter={{ id: encounter.encounterId, tcaUtc: encounter.tcaUtc, objectAId: encounter.objectAId, objectBId: encounter.objectBId }}
        palette={palette}
        onSelect={setSelected}
      />
      <TimelineScrubber
        value={index}
        onChange={setIndex}
        playing={playing}
        onPlayingChange={setPlaying}
        speed={speed}
        onSpeedChange={setSpeed}
        tcaLabel={encounter.tcaUtc.slice(11, 19)}
        utcLabel={utcLabel}
        halfWindowMinutes={halfWindowMinutes}
        intervals={intervals}
      />
      <p>Sample separation {sample.separationMeters.toFixed(1)} m at {utcLabel}</p>
      {comparison && <p>What-if sample separation {comparison.separationMeters.toFixed(1)} m at {utcLabel}</p>}
      <SeparationChart
        current={index / intervals * 100}
        minimumMeters={encounter.minimumSeparationMeters}
        samples={baselineSeparation}
        comparisonSamples={comparisonSeparation}
        replayHalfWindowMinutes={halfWindowMinutes}
      />
      <p>Source {source.sourceName} · Snapshot {source.snapshotId} · Retrieved {source.retrievedAtUtc} · Older element epoch {encounter.sourceEpochUtc}</p>
      <p>{source.caveat} {trajectory.frame.frameCaveat}</p>
      <section className="screening-results" aria-label="Screened encounter what-if">
        <h4>Educational what-if</h4>
        <p>Apply a hypothetical change parallel to {firstName}’s velocity at the replay start. The model estimates the change using central Earth gravity.</p>
        <form className="screening-form" onSubmit={(event) => { void preview(event); }}>
          <div className="screening-fields">
            <label>Along-track delta-v (m/s)
              <input type="number" min="-0.5" max="0.5" step="any" required value={deltaV} disabled={submitting}
                onChange={(event) => setDeltaV(event.currentTarget.value)} />
            </label>
          </div>
          <button type="submit" className="primary-button screening-button" disabled={submitting}>
            {submitting ? 'Computing what-if…' : 'Preview what-if'}
          </button>
        </form>
        {scenarioError && <p className="screening-error" role="alert">{scenarioError}</p>}
        {scenario && comparisonSeparation && (
          <>
            <p role="status">Preview applied at {scenario.impulseAtUtc} with {scenario.alongTrackDeltaVMps} m/s.</p>
            <p>Minimum across replay samples · Baseline {Math.min(...baselineSeparation).toFixed(1)} m · What-if {Math.min(...comparisonSeparation).toFixed(1)} m</p>
            <ul className="screening-caveats">{scenario.caveats.map((caveat) => <li key={caveat}>{caveat}</li>)}</ul>
            <button type="button" className="text-button" disabled={submitting} onClick={() => { setScenario(null); setScenarioError(''); }}>
              Clear what-if
            </button>
          </>
        )}
      </section>
    </section>
  );
}
