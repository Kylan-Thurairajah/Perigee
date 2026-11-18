import { useEffect, useMemo, useState } from 'react';
import type { SceneMode, SceneSelection } from '../components/SceneSlot';
import SceneSlot from '../components/SceneSlot';
import SeparationChart from '../components/SeparationChart';
import TimelineScrubber from '../components/TimelineScrubber';
import { createFixtureOrbitPaths, createFixtureSeparationSamples } from '../data/fixtureOrbitPaths';
import { catalogObjects, catalogFixture, featuredEncounter, getObject } from '../data/fixtures';

type View = 'explorer' | 'encounter' | 'maneuver';
const THEME_KEY = 'perigee-theme';

function getInitialTheme(): 'light' | 'dark' {
  try {
    return localStorage.getItem(THEME_KEY) === 'dark' ? 'dark' : 'light';
  } catch {
    return 'light';
  }
}

function formatUtc(value: string) {
  return new Intl.DateTimeFormat('en-GB', {
    timeZone: 'UTC',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  }).format(new Date(value));
}

function App() {
  const [theme, setTheme] = useState<'light' | 'dark'>(getInitialTheme);
  const [view, setView] = useState<View>('explorer');
  const [selected, setSelected] = useState<SceneSelection>(null);
  const [timelineValue, setTimelineValue] = useState(50);
  const [playing, setPlaying] = useState(false);
  const [speed, setSpeed] = useState(1);
  const [deltaV, setDeltaV] = useState(0.12);
  const [recalculated, setRecalculated] = useState(false);
  const encounter = featuredEncounter;
  const objectA = getObject(encounter.objectAId)!;
  const objectB = getObject(encounter.objectBId)!;
  const orbitPaths = useMemo(() => createFixtureOrbitPaths(catalogObjects, encounter), [encounter]);
  const separationSamples = useMemo(
    () => createFixtureSeparationSamples(orbitPaths, encounter.objectAId, encounter.objectBId),
    [encounter.objectAId, encounter.objectBId, orbitPaths],
  );

  const sceneMode: SceneMode = view === 'explorer' ? 'explore' : view;
  const sceneTimeUtc = useMemo(() => {
    const tca = new Date(encounter.tcaUtc).getTime();
    const offsetMinutes = (timelineValue - 50) * 0.6;
    return new Date(tca + offsetMinutes * 60_000).toISOString();
  }, [encounter.tcaUtc, timelineValue]);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch {
      // Theme remains usable when browser storage is unavailable.
    }
  }, [theme]);

  useEffect(() => {
    if (selected?.kind === 'encounter' && view === 'explorer') setView('encounter');
  }, [selected, view]);

  function openEncounter() {
    setSelected({ kind: 'encounter', id: encounter.encounterId });
    setView('encounter');
  }

  function goTo(nextView: View) {
    setView(nextView);
    setSelected(nextView === 'explorer' ? null : { kind: 'encounter', id: encounter.encounterId });
  }

  const navItems: { id: View; label: string }[] = [
    { id: 'explorer', label: 'Explorer' },
    { id: 'encounter', label: 'Encounter' },
    { id: 'maneuver', label: 'Maneuver lab' },
  ];

  return (
    <main className="app-shell">
      <header className="topbar">
        <button className="brand" onClick={() => goTo('explorer')} aria-label="Perigee home">
          <span className="brand-mark" aria-hidden="true"><i /><i /><i /></span>
          <span>perigee<span className="brand-period">.</span></span>
        </button>
        <nav className="primary-nav" aria-label="Main navigation">
          {navItems.map((item) => (
            <button key={item.id} className={view === item.id ? 'nav-active' : ''} onClick={() => goTo(item.id)} aria-current={view === item.id ? 'page' : undefined}>
              {item.label}
            </button>
          ))}
        </nav>
        <div className="topbar-right">
          <span className="snapshot-pill"><span /> FIXTURE SNAPSHOT</span>
          <button
            className="theme-toggle"
            onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
            aria-label={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
            title={`Switch to ${theme === 'light' ? 'dark' : 'light'} theme`}
          >
            {theme === 'light' ? <span aria-hidden="true">☾</span> : <span aria-hidden="true">☼</span>}
          </button>
          <button className="avatar" aria-label="Demo workspace">P</button>
        </div>
      </header>

      <div className="page-content">
        <div className="page-heading">
          <div>
            <div className="eyebrow"><span className="eyebrow-line" /> ORBITAL TRAFFIC, MADE VISIBLE</div>
            <h1>{view === 'explorer' ? 'A little more perspective.' : view === 'encounter' ? 'A closer look.' : 'What if?'}</h1>
            <p>{view === 'explorer' ? 'Explore the movement and moments shaping our near-Earth neighborhood.' : view === 'encounter' ? 'Replay a fixture close approach, one moment at a time.' : 'Explore a simplified, hypothetical along-track change.'}</p>
          </div>
          <div className="date-stamp"><span className="calendar-icon">◷</span><div><strong>20 MAR 2026</strong><small>FIXTURE EPOCH · 12:00 UTC</small></div></div>
        </div>

        <div className={`workspace workspace-${view}`}>
          <div className="visual-column">
            <SceneSlot
              mode={sceneMode}
              sceneTimeUtc={sceneTimeUtc}
              selected={selected}
              satellites={catalogObjects.map((item) => ({ id: item.catalogId, name: item.name, altitudeKm: item.meanAltitudeKm }))}
              paths={orbitPaths}
              encounter={{ id: encounter.encounterId, tcaUtc: encounter.tcaUtc, objectAId: encounter.objectAId, objectBId: encounter.objectBId }}
              palette={theme}
              onSelect={(selection) => {
                setSelected(selection);
                if (selection?.kind === 'encounter') setView('encounter');
              }}
            />
            <TimelineScrubber
              value={timelineValue}
              onChange={setTimelineValue}
              playing={playing}
              onPlayingChange={setPlaying}
              speed={speed}
              onSpeedChange={setSpeed}
              tcaLabel={formatUtc(encounter.tcaUtc)}
              utcLabel={sceneTimeUtc}
            />
          </div>

          {view === 'explorer' && (
            <aside className="side-panel explorer-panel" aria-label="Explorer details">
              <div className="panel-title-row"><div><span className="section-index">01 / OVERVIEW</span><h2>In your orbit</h2></div><button className="icon-button" aria-label="More overview options">···</button></div>
              <p className="panel-intro">A small sample of objects moving through low Earth orbit.</p>
              <div className="object-list">
                {catalogObjects.slice(0, 3).map((object, index) => (
                  <button
                    className={`object-row ${selected?.kind === 'satellite' && selected.id === object.catalogId ? 'object-selected' : ''}`}
                    key={object.catalogId}
                    onClick={() => setSelected({ kind: 'satellite', id: object.catalogId })}
                  >
                    <span className={`object-icon object-icon-${index}`}><i /></span>
                    <span className="object-name">{object.name}<small>{object.orbitRegime.toUpperCase()} · {object.meanAltitudeKm} KM</small></span>
                    <span className="object-status" aria-label="Fixture object">●</span>
                  </button>
                ))}
              </div>
              <div className="panel-divider" />
              <button className="featured-card" onClick={openEncounter}>
                <div className="featured-top"><span className="featured-marker"><i /></span><span>FEATURED ENCOUNTER</span><span className="arrow-up">↗</span></div>
                <h3>{objectA.name}<span> × </span>{objectB.name}</h3>
                <div className="featured-metrics"><span><small>MISS DISTANCE</small><strong>{encounter.minimumSeparationMeters.toLocaleString()} <em>m</em></strong></span><span><small>TCA · 20 MAR</small><strong>{formatUtc(encounter.tcaUtc)} <em>UTC</em></strong></span></div>
                <span className="open-analysis">Open analysis <span>→</span></span>
              </button>
              <div className="fixture-note"><span className="info-icon">i</span><p>Fictional, deterministic fixture data. Not operational ephemerides or collision predictions.</p></div>
            </aside>
          )}

          {view === 'encounter' && (
            <aside className="side-panel analysis-panel" aria-label="Encounter analysis">
              <div className="panel-title-row"><div><span className="section-index">ENCOUNTER · SYNTHETIC</span><h2>{objectA.name} <span className="title-cross">×</span> {objectB.name}</h2></div><span className="risk-tag"><i /> CLOSE APPROACH</span></div>
              <div className="analysis-time"><span className="time-icon">◷</span><div><strong>20 March 2026 · {formatUtc(encounter.tcaUtc)} UTC</strong><small>Time of closest approach</small></div></div>
              <div className="metric-row">
                <div className="metric-card"><span>MIN SEPARATION</span><strong>{encounter.minimumSeparationMeters.toLocaleString()}<small> m</small></strong></div>
                <div className="metric-card"><span>RELATIVE SPEED</span><strong>{(encounter.relativeSpeedMps / 1000).toFixed(2)}<small> km/s</small></strong></div>
              </div>
              <div className="chart-card"><div className="chart-heading"><div><span className="section-index">REPLAY PROFILE</span><h3>Separation over time</h3></div><span className="chart-unit">METERS</span></div><SeparationChart current={timelineValue} minimumMeters={encounter.minimumSeparationMeters} samples={separationSamples} /></div>
              <div className="pair-card">
                <span className="section-index">OBJECT PAIR</span>
                {[objectA, objectB].map((object, index) => <div className="pair-row" key={object.catalogId}><span className={`pair-orbit pair-orbit-${index}`} /><span>{object.name}<small>{object.meanAltitudeKm} km · LEO</small></span><span className="pair-id">{object.catalogId.replace('sat-', '').toUpperCase()}</span></div>)}
              </div>
              <button className="primary-button maneuver-button" onClick={() => goTo('maneuver')}>Explore a what-if <span>↗</span></button>
              <p className="caveat-copy">Separation is not collision probability. Paths and replay are simplified synthetic fixtures, not SGP4 propagation.</p>
            </aside>
          )}

          {view === 'maneuver' && (
            <aside className="side-panel maneuver-panel" aria-label="Maneuver lab">
              <div className="panel-title-row"><div><span className="section-index">MANEUVER LAB · EDUCATIONAL</span><h2>A small what-if.</h2></div><span className="lab-status"><i /> PREVIEW</span></div>
              <p className="panel-intro">Adjust an along-track impulse and compare the illustrative scenario.</p>
              <div className="maneuver-control">
                <div className="control-label"><span>Along-track Δv</span><strong>{deltaV.toFixed(2)} <small>m/s</small></strong></div>
                <input aria-label="Along-track delta-v in meters per second" type="range" min="-0.5" max="0.5" step="0.01" value={deltaV} onChange={(event) => { setDeltaV(Number(event.target.value)); setRecalculated(false); }} />
                <div className="range-labels"><span>−0.5 m/s</span><span>+0.5 m/s</span></div>
              </div>
              <div className="maneuver-at"><span className="section-index">SIMULATED IMPULSE</span><strong>20 Mar 2026 · 14:02 UTC</strong><small>30 minutes before closest approach</small></div>
              <div className="comparison-card">
                <span className="section-index">{recalculated ? 'LOCAL PREVIEW UPDATED' : 'ORIGINAL VS. PREVIEW'}</span>
                <div className="comparison-row"><span><i className="legend-dot original" />Original fixture</span><strong>{encounter.minimumSeparationMeters.toLocaleString()} m</strong></div>
                <div className="comparison-row"><span><i className="legend-dot hypothetical" />What-if preview</span><strong>{Math.round(encounter.minimumSeparationMeters + Math.abs(deltaV) * 1120).toLocaleString()} m</strong></div>
              </div>
              <button className="primary-button recalculate-button" onClick={() => setRecalculated(true)}>{recalculated ? 'Preview updated' : 'Preview scenario'} <span>→</span></button>
              <div className="limitations"><span className="info-icon">i</span><p><strong>Simplified preview, not a propagation.</strong> This local illustration does not model orbital dynamics, uncertainty, or collision probability. Never use as operational guidance.</p></div>
              <button className="text-button" onClick={() => { setDeltaV(0.12); setRecalculated(false); }}>Reset scenario <span>↺</span></button>
            </aside>
          )}
        </div>
        <footer className="page-footer"><span>PERIGEE <i>·</i> A CLEARER VIEW OF WHAT’S HAPPENING IN ORBIT</span><span>LOCAL DEMO <b>·</b> {catalogFixture.catalogSnapshotId.toUpperCase()}</span></footer>
      </div>
    </main>
  );
}

export default App;
