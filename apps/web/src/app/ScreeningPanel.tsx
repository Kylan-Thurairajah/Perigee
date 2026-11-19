import { useEffect, useState } from 'react';
import { requestJSON } from './screeningApi';
import ScreeningReplay, { type ScreeningEncounterDetail } from './ScreeningReplay';

type ImportedRecord = {
  noradCatId: number;
  propagationSupported: boolean;
  dataQuality: string;
};

type OMMImportResponse = {
  snapshotId: string;
  records: ImportedRecord[];
  caveats: string[];
};

type ScreeningStatus = {
  runId: string;
  status: 'queued' | 'running' | 'completed' | 'failed' | 'cancelled';
  phase: string;
  progressPercent: number;
  error?: { message: string };
};

export type ScreeningEncounter = {
  encounterId: string;
  objectAId: string;
  objectBId: string;
  tcaUtc: string;
  minimumSeparationMeters: number;
  relativeSpeedMps: number;
};

type EncounterPage = {
  items: ScreeningEncounter[];
  nextCursor: string | null;
  caveats: string[];
};

function toLocalDateTime(value: Date) {
  const local = new Date(value.getTime() - value.getTimezoneOffset() * 60_000);
  return local.toISOString().slice(0, 16);
}

function toUTC(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) throw new Error('Enter a valid date and time.');
  return parsed.toISOString();
}

function formatUTC(value: string) {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'UTC',
    dateStyle: 'medium',
    timeStyle: 'medium',
  }).format(new Date(value));
}

function ScreeningPanel({ visible, palette }: { visible: boolean; palette: 'light' | 'dark' }) {
  const [file, setFile] = useState<File | null>(null);
  const [sourceName, setSourceName] = useState('');
  const [snapshotId, setSnapshotId] = useState(() => `local-omm-${Date.now()}`);
  const [retrievedAt, setRetrievedAt] = useState(() => toLocalDateTime(new Date()));
  const [maxEpochAgeDays, setMaxEpochAgeDays] = useState(7);
  const [startTime, setStartTime] = useState(() => toLocalDateTime(new Date()));
  const [endTime, setEndTime] = useState(() => toLocalDateTime(new Date(Date.now() + 60 * 60_000)));
  const [thresholdMeters, setThresholdMeters] = useState(10_000);
  const [coarseStepSeconds, setCoarseStepSeconds] = useState(60);
  const [imported, setImported] = useState<OMMImportResponse | null>(null);
  const [activeRunId, setActiveRunId] = useState('');
  const [status, setStatus] = useState<ScreeningStatus | null>(null);
  const [results, setResults] = useState<ScreeningEncounter[] | null>(null);
  const [caveats, setCaveats] = useState<string[]>([]);
  const [nextCursor, setNextCursor] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState('');
  const [pollError, setPollError] = useState('');
  const [pollAttempt, setPollAttempt] = useState(0);
  const [selectedEncounter, setSelectedEncounter] = useState<string | null>(null);
  const [detail, setDetail] = useState<ScreeningEncounterDetail | null>(null);
  const [detailError, setDetailError] = useState('');
  const [detailAttempt, setDetailAttempt] = useState(0);
  const supportedCount = imported?.records.filter((record) => record.propagationSupported && record.dataQuality !== 'stale').length ?? 0;
  const canScreen = imported !== null && supportedCount >= 2 && supportedCount <= 100 && supportedCount === imported.records.length;
  const runPending = status?.status === 'queued' || status?.status === 'running';

  useEffect(() => {
    if (!activeRunId) return undefined;
    let active = true;
    let timer: number | undefined;

    async function poll() {
      try {
        const nextStatus = await requestJSON<ScreeningStatus>(`/api/screenings/${activeRunId}`);
        if (!active) return;
        setStatus(nextStatus);
        if (nextStatus.status === 'failed' || nextStatus.status === 'cancelled') {
          setError(nextStatus.error?.message || `Screening ${nextStatus.status}.`);
          return;
        }
        if (nextStatus.status === 'completed') {
          const page = await requestJSON<EncounterPage>(`/api/screenings/${activeRunId}/encounters`);
          if (!active) return;
          setResults(page.items);
          setCaveats(page.caveats);
          setNextCursor(page.nextCursor);
          return;
        }
        timer = window.setTimeout(() => { void poll(); }, 500);
      } catch (requestError) {
        if (active) setPollError(requestError instanceof Error ? requestError.message : 'Could not read screening status.');
      }
    }

    void poll();
    return () => {
      active = false;
      if (timer !== undefined) window.clearTimeout(timer);
    };
  }, [activeRunId, pollAttempt]);

  useEffect(() => {
    if (!activeRunId || !selectedEncounter) return;
    const controller = new AbortController();
    const url = `/api/screenings/${encodeURIComponent(activeRunId)}/encounters/${encodeURIComponent(selectedEncounter)}`;
    requestJSON<ScreeningEncounterDetail>(url, { signal: controller.signal }).then(
      (data) => { if (!controller.signal.aborted) setDetail(data); },
      (failure: unknown) => {
        if (!controller.signal.aborted) setDetailError(failure instanceof Error ? failure.message : 'Could not load encounter replay.');
      },
    );
    return () => controller.abort();
  }, [activeRunId, selectedEncounter, detailAttempt]);

  function selectEncounter(id: string | null) {
    if (id === selectedEncounter) return;
    setSelectedEncounter(id);
    setDetail(null);
    setDetailError('');
  }

  async function importSnapshot(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!file || busy || loadingMore) return;
    if (file.size > 4 * 1024 * 1024) {
      setError('Choose an OMM JSON file smaller than 4 MiB.');
      return;
    }
    setBusy(true);
    setError('');
    setPollError('');
    setCaveats([]);
    setNextCursor(null);
    setImported(null);
    selectEncounter(null);
    setStatus(null);
    setResults(null);
    setActiveRunId('');
    try {
      const records: unknown = JSON.parse(await file.text());
      const result = await requestJSON<OMMImportResponse>('/api/catalog/omm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          source: {
            name: sourceName.trim(),
            snapshotId: snapshotId.trim(),
            retrievedAtUtc: toUTC(retrievedAt),
          },
          maxEpochAgeSeconds: maxEpochAgeDays * 86_400,
          records,
        }),
      });
      setImported(result);
      setSnapshotId(result.snapshotId);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'OMM import failed.');
    } finally {
      setBusy(false);
    }
  }

  async function startScreening(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!canScreen || busy || loadingMore || (runPending && !pollError)) return;
    setBusy(true);
    setError('');
    setPollError('');
    setActiveRunId('');
    setResults(null);
    selectEncounter(null);
    setCaveats([]);
    setNextCursor(null);
    setStatus({ runId: '', status: 'queued', phase: 'queued', progressPercent: 0 });
    try {
      const created = await requestJSON<{ runId: string }>('/api/screenings', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          catalogSnapshotId: imported.snapshotId,
          startTimeUtc: toUTC(startTime),
          endTimeUtc: toUTC(endTime),
          thresholdMeters,
          coarseStepSeconds,
        }),
      });
      setActiveRunId(created.runId);
    } catch (requestError) {
      setStatus(null);
      setError(requestError instanceof Error ? requestError.message : 'Could not start screening.');
    } finally {
      setBusy(false);
    }
  }

  async function loadMore() {
    if (!activeRunId || !nextCursor) return;
    setLoadingMore(true);
    setError('');
    try {
      const page = await requestJSON<EncounterPage>(
        `/api/screenings/${activeRunId}/encounters?cursor=${encodeURIComponent(nextCursor)}`,
      );
      setResults((current) => [...(current ?? []), ...page.items]);
      setNextCursor(page.nextCursor);
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : 'Could not load more results.');
    } finally {
      setLoadingMore(false);
    }
  }

  return (
    <section className="side-panel screening-panel" aria-label="Local OMM screening">
      <div className="panel-title-row">
        <div><span className="section-index">LOCAL DATA · SGP4 SCREENING</span><h2>Screen a snapshot</h2></div>
        <span className="lab-status"><i /> IN MEMORY</span>
      </div>
      <p className="panel-intro">Import an OMM JSON file from this device, then run a bounded close-approach screen. The snapshot is not uploaded to a remote service or saved by the API.</p>

      <form className="screening-form" onSubmit={(event) => { void importSnapshot(event); }}>
        <label className="screening-file-field">OMM JSON file
          <input type="file" accept=".json,application/json" onChange={(event) => setFile(event.currentTarget.files?.[0] ?? null)} />
        </label>
        <div className="screening-fields">
          <label>Source name<input required value={sourceName} onChange={(event) => setSourceName(event.currentTarget.value)} placeholder="e.g. Local OMM snapshot" /></label>
          <label>Snapshot ID<input required value={snapshotId} onChange={(event) => setSnapshotId(event.currentTarget.value)} /></label>
          <label>Retrieved at (local time)<input required type="datetime-local" value={retrievedAt} onChange={(event) => setRetrievedAt(event.currentTarget.value)} /></label>
          <label>Maximum element age (days)<input required type="number" min="1" step="1" value={maxEpochAgeDays} onChange={(event) => setMaxEpochAgeDays(Number(event.currentTarget.value))} /></label>
        </div>
        <button className="primary-button screening-button" type="submit" disabled={busy || loadingMore || !file}>
          {busy && !imported ? 'Validating snapshot…' : 'Validate and import OMM'} <span>↓</span>
        </button>
      </form>

      {imported && (
        <div className="screening-import-summary" role="status">
          <strong>{imported.records.length} records received · {supportedCount} supported and fresh</strong>
          <span>Snapshot <code>{imported.snapshotId}</code> {canScreen ? 'is ready for screening.' : 'requires 2–100 records, all fresh and supported, to run screening.'}</span>
          {imported.caveats.map((caveat) => <span key={caveat}>{caveat}</span>)}
        </div>
      )}

      <form className="screening-form screening-run-form" onSubmit={(event) => { void startScreening(event); }}>
        <div className="screening-section-heading"><span className="section-index">SCREENING WINDOW</span><span>Max 100 objects · 1,441 samples</span></div>
        <div className="screening-fields">
          <label>Start time (local)<input required type="datetime-local" value={startTime} onChange={(event) => setStartTime(event.currentTarget.value)} /></label>
          <label>End time (local)<input required type="datetime-local" value={endTime} onChange={(event) => setEndTime(event.currentTarget.value)} /></label>
          <label>Distance threshold (m)<input required type="number" min="0.001" step="any" value={thresholdMeters} onChange={(event) => setThresholdMeters(Number(event.currentTarget.value))} /></label>
          <label>Coarse step (seconds)<input required type="number" min="1" max="3600" step="1" value={coarseStepSeconds} aria-describedby="screening-step-help" onChange={(event) => setCoarseStepSeconds(Number(event.currentTarget.value))} /></label>
        </div>
        <p id="screening-step-help" className="caveat-copy">Larger steps can miss close approaches. Smaller steps use more of the 1,441-sample budget.</p>
        <button className="primary-button screening-button" type="submit" disabled={!canScreen || busy || loadingMore || (runPending && !pollError)}>
          {busy && imported ? 'Submitting…' : 'Run screening'} <span>→</span>
        </button>
      </form>

      {error && <p className="screening-error" role="alert">{error}</p>}

      {pollError && (
        <>
          <p className="screening-error" role="alert">{pollError}</p>
          <button className="text-button" type="button" onClick={() => {
            setPollError('');
            setPollAttempt((attempt) => attempt + 1);
          }}>Retry status and results</button>
        </>
      )}

      {status && (
        <div className="screening-status" role="status" aria-live="polite">
          <div><strong>{status.status === 'completed' ? 'Screening complete' : status.status === 'failed' || status.status === 'cancelled' ? `Screening ${status.status}` : 'Screening in progress'}</strong>
            <span>{status.phase} · {status.progressPercent}%</span></div>
          <progress max="100" value={status.progressPercent} aria-label="Screening progress" />
        </div>
      )}

      {status?.status === 'completed' && results === null && !pollError && <p role="status">Loading screening results…</p>}
      {results !== null && (
        <div className="screening-results">
          <div className="screening-section-heading"><span className="section-index">CLOSE APPROACH RESULTS</span><span>{results.length}{nextCursor ? '+' : ''} shown</span></div>
          {results.length === 0 ? <p className="screening-empty">No encounters were found within this threshold.</p> : (
            <div className="screening-table-wrap">
              <table>
                <thead><tr><th>Object pair</th><th>TCA · UTC</th><th>Miss distance</th><th>Relative speed</th><th>Replay</th></tr></thead>
                <tbody>{results.map((result) => (
                  <tr key={result.encounterId}>
                    <td>{result.objectAId.replace('omm-norad-', 'NORAD ')} × {result.objectBId.replace('omm-norad-', 'NORAD ')}</td>
                    <td>{formatUTC(result.tcaUtc)}</td>
                    <td>{result.minimumSeparationMeters.toLocaleString(undefined, { maximumFractionDigits: 1 })} m</td>
                    <td>{result.relativeSpeedMps.toLocaleString(undefined, { maximumFractionDigits: 1 })} m/s</td>
                    <td><button className="text-button" type="button" aria-pressed={selectedEncounter === result.encounterId} aria-label={`Replay ${result.objectAId} and ${result.objectBId}`} onClick={() => selectEncounter(result.encounterId)}>Open replay</button></td>
                  </tr>
                ))}</tbody>
              </table>
            </div>
          )}
          {nextCursor && <button className="text-button" type="button" disabled={loadingMore} onClick={() => { void loadMore(); }}>{loadingMore ? 'Loading…' : 'Load more results'} <span>↓</span></button>}
          <ul className="screening-caveats">{caveats.map((caveat) => <li key={caveat}>{caveat}</li>)}</ul>
        </div>
      )}
      {selectedEncounter && (
        <div className="screening-results">
          <button className="text-button" type="button" onClick={() => selectEncounter(null)}>Close replay</button>
          {detailError && (
            <>
              <p className="screening-error" role="alert">{detailError}</p>
              <button className="text-button" type="button" onClick={() => {
                setDetailError('');
                setDetailAttempt((attempt) => attempt + 1);
              }}>Retry replay</button>
            </>
          )}
          {!detailError && !detail && <p role="status">Loading encounter replay…</p>}
          {visible && detail && <ScreeningReplay key={selectedEncounter} detail={detail} palette={palette} runId={activeRunId} />}
        </div>
      )}
    </section>
  );
}

export default ScreeningPanel;
