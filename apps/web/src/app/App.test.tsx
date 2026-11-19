import { act, cleanup, fireEvent, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('../scene/OrbitalScene', () => ({
  default: () => <div data-testid="orbital-scene" />,
}));

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
});

describe('Perigee application', () => {
  it.each([
    ['0.5×', 8000],
    ['1×', 4000],
    ['2×', 2000],
  ])('triggers the collision at the correct replay time at %s', (speed, impactMs) => {
    vi.useFakeTimers({ toFake: ['requestAnimationFrame', 'cancelAnimationFrame', 'performance'] });
    render(<App />);
    fireEvent.click(screen.getByRole('button', { name: 'Encounter' }));
    fireEvent.click(screen.getByRole('button', { name: speed }));
    fireEvent.click(screen.getByRole('button', { name: 'Play fictional collision' }));
    const scene = screen.getByRole('region', { name: 'Orbital scene preview' });
    act(() => vi.advanceTimersByTime(impactMs - 32));
    expect(scene).toHaveAttribute('data-collision-phase', 'approach');
    act(() => vi.advanceTimersByTime(64));
    expect(scene).toHaveAttribute('data-collision-phase', 'impact');
    const effectMs = impactMs / 2;
    act(() => vi.advanceTimersByTime(effectMs));
    expect(scene).toHaveAttribute('data-collision-phase', 'aftermath');
    act(() => vi.advanceTimersByTime(effectMs));
    expect(scene).toHaveAttribute('data-collision-phase', 'approach');
  });

  it('pauses and scrubs the impact, preserves position on speed changes, and restarts', () => {
    vi.useFakeTimers({ toFake: ['requestAnimationFrame', 'cancelAnimationFrame', 'performance'] });
    render(<App />);
    fireEvent.click(screen.getByRole('button', { name: 'Encounter' }));
    fireEvent.click(screen.getByRole('button', { name: 'Play fictional collision' }));
    const scene = screen.getByRole('region', { name: 'Orbital scene preview' });
    const timeline = screen.getByRole('slider', { name: 'Encounter replay time' });
    act(() => vi.advanceTimersByTime(1008));
    const beforeChange = (timeline as HTMLInputElement).valueAsNumber;
    fireEvent.click(screen.getByRole('button', { name: '2×' }));
    expect((timeline as HTMLInputElement).valueAsNumber).toBe(beforeChange);
    act(() => vi.advanceTimersByTime(800));
    expect((timeline as HTMLInputElement).valueAsNumber - beforeChange).toBeCloseTo(20);

    fireEvent.click(screen.getByRole('button', { name: 'Pause encounter replay' }));
    fireEvent.change(timeline, { target: { value: '50' } });
    expect(scene).toHaveAttribute('data-collision-phase', 'impact');
    expect(screen.getByText('2 OBJECTS IN VIEW')).toBeInTheDocument();
    expect(scene).toHaveAttribute('data-scene-time', '2026-03-20T14:32:10.000Z');
    act(() => vi.advanceTimersByTime(4000));
    expect(scene).toHaveAttribute('data-collision-phase', 'impact');
    fireEvent.change(timeline, { target: { value: '75' } });
    expect(scene).toHaveAttribute('data-collision-phase', 'aftermath');
    fireEvent.change(timeline, { target: { value: '25' } });
    expect(scene).toHaveAttribute('data-collision-phase', 'approach');
    expect(screen.getByText('4 OBJECTS IN VIEW')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Restart collision replay' }));
    expect(timeline).toHaveValue('0');
    expect(screen.getByRole('button', { name: 'Pause encounter replay' })).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Back to close approach' }));
    expect(scene).not.toHaveAttribute('data-collision-phase');
    expect(screen.getByText('742')).toBeInTheDocument();
  });

  it('imports a snapshot, pages through results, and replays a computed encounter', async () => {
    const user = userEvent.setup();
    const records = [101, 102].map((noradCatId) => ({ noradCatId, propagationSupported: true, dataQuality: 'fresh' }));
    const encounter = {
      encounterId: 'first', objectAId: 'omm-norad-101', objectBId: 'omm-norad-102',
      tcaUtc: '2026-03-20T14:32:10Z', minimumSeparationMeters: 123, relativeSpeedMps: 456,
    };
    const completed = { runId: 'run-1', status: 'completed', phase: 'complete', progressPercent: 100 };
    const tca = Date.parse(encounter.tcaUtc);
    const samples = Array.from({ length: 61 }, (_, index) => {
      const separationMeters = 123 + Math.abs(index - 30);
      return {
        atUtc: new Date(tca + (index - 30) * 10_000).toISOString(),
        objectA: { positionXKm: 7000, positionYKm: 0, positionZKm: 0 },
        objectB: { positionXKm: 7000, positionYKm: separationMeters / 1000, positionZKm: 0 },
        separationMeters,
      };
    });
    const fetchMock = vi.fn()
      .mockResolvedValueOnce(Response.json({ snapshotId: 'local-test', records, caveats: [] }))
      .mockResolvedValueOnce(Response.json({ runId: 'run-1' }))
      .mockResolvedValueOnce(Response.json({ ...completed, status: 'running', progressPercent: 10 }))
      .mockResolvedValueOnce(Response.json(completed))
      .mockResolvedValueOnce(new Response('Bad Gateway', { status: 502 }))
      .mockResolvedValueOnce(Response.json(completed))
      .mockResolvedValueOnce(Response.json({ items: [encounter], nextCursor: 'next+page', caveats: ['Screening caveat'] }))
      .mockResolvedValueOnce(Response.json({ items: [{ ...encounter, encounterId: 'second', minimumSeparationMeters: 234 }], nextCursor: null, caveats: [] }))
      .mockResolvedValueOnce(Response.json({
        encounter: { ...encounter, sourceEpochUtc: '2026-03-19T00:00:00Z' },
        objectAName: 'Object A', objectBName: 'Object B',
        trajectory: {
          frame: { frame: 'TEME', frameCaveat: 'Illustrative Earth.' },
          startUtc: samples[0].atUtc, endUtc: samples[60].atUtc, samples,
          source: { sourceName: 'Local test', snapshotId: 'local-test', retrievedAtUtc: '2026-03-20T00:00:00Z', caveat: 'Bounded replay.' },
        },
      }))
      .mockResolvedValueOnce(Response.json({
        impulseAtUtc: samples[0].atUtc, alongTrackDeltaVMps: 0.1, caveats: ['Educational comparison.'],
        samples: samples.map((sample, index) => ({ ...sample, separationMeters: sample.separationMeters + index })),
      }));
    vi.stubGlobal('fetch', fetchMock);
    render(<App />);

    await user.click(screen.getByRole('button', { name: 'Screening' }));
    const file = new File(['[]'], 'catalog.json', { type: 'application/json' });
    Object.defineProperty(file, 'text', { value: async () => '[]' });
    await user.upload(screen.getByLabelText('OMM JSON file'), file);
    await user.type(screen.getByLabelText('Source name'), 'Local test');
    await user.click(screen.getByRole('button', { name: /validate and import omm/i }));
    expect(await screen.findByText(/2 supported and fresh/)).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /run screening/i }));

    expect(await screen.findByRole('alert')).toHaveTextContent('Request failed (502)');
    expect(screen.queryByText(/no encounters were found/i)).toBeNull();
    await user.click(screen.getByRole('button', { name: 'Retry status and results' }));
    expect(await screen.findByText('123 m')).toBeInTheDocument();
    expect(screen.getByText('Screening caveat')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: /load more results/i }));
    expect(await screen.findByText('234 m')).toBeInTheDocument();
    expect(fetchMock).toHaveBeenLastCalledWith('/api/screenings/run-1/encounters?cursor=next%2Bpage', undefined);

    await user.click(screen.getByRole('button', { name: 'Explorer' }));
    expect(screen.queryByRole('region', { name: 'Local OMM screening' })).toBeNull();
    await user.click(screen.getByRole('button', { name: 'Screening' }));
    expect(screen.getByText('234 m')).toBeInTheDocument();
    await user.click(screen.getAllByRole('button', { name: /Replay omm-norad/i })[0]);
    expect(await screen.findByRole('heading', { name: 'Object A × Object B' })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /SGP4 separation profile/i })).toBeInTheDocument();
    const scene = screen.getByRole('region', { name: 'Orbital scene preview' });
    expect(scene).toHaveAttribute('data-path-count', '2');
    expect(scene).toHaveAttribute('data-scene-time', samples[30].atUtc);
    fireEvent.change(screen.getByRole('slider', { name: /encounter replay time/i }), { target: { value: '31' } });
    expect(scene).toHaveAttribute('data-scene-time', samples[31].atUtc);
    expect(screen.getByText('Sample separation 124.0 m at 14:32:20 UTC')).toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: 'Preview what-if' }));
    expect(await screen.findByText('What-if sample separation 155.0 m at 14:32:20 UTC')).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /two-body what-if overlay/i })).toBeInTheDocument();
    expect(scene).toHaveAttribute('data-path-count', '3');
    expect(fetchMock).toHaveBeenLastCalledWith('/api/screenings/run-1/encounters/first/what-if', expect.objectContaining({
      method: 'POST', body: '{"alongTrackDeltaVMps":0.1}',
    }));
    await user.click(screen.getByRole('button', { name: 'Clear what-if' }));
    expect(scene).toHaveAttribute('data-path-count', '2');
  });

  it('opens the featured fixture encounter and its replay analysis', async () => {
    const user = userEvent.setup();
    render(<App />);

    expect(screen.getByTestId('orbital-scene')).toBeInTheDocument();
    expect(screen.getByRole('slider', { name: /encounter replay time/i })).toHaveValue('50');
    expect(screen.getByRole('region', { name: /orbital scene preview/i })).toHaveAttribute('data-path-count', '4');
    expect(screen.getByRole('region', { name: /orbital scene preview/i })).toHaveAttribute('data-scene-time', '2026-03-20T14:32:10.000Z');
    await user.click(screen.getByRole('button', { name: /featured encounter/i }));

    expect(screen.getByRole('heading', { name: /Peregrine A.*Lyra B/i })).toBeInTheDocument();
    expect(screen.getByText('742')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: /encounter timeline/i })).toBeInTheDocument();
    expect(screen.getByRole('img', { name: /synthetic separation profile/i })).toBeInTheDocument();
  });

  it('toggles and persists the theme preference', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole('button', { name: 'Switch to dark theme' }));

    expect(document.documentElement).toHaveAttribute('data-theme', 'dark');
    expect(localStorage.getItem('perigee-theme')).toBe('dark');
  });

  it('provides a bounded illustrative maneuver preview', async () => {
    const user = userEvent.setup();
    render(<App />);

    await user.click(screen.getByRole('button', { name: 'Maneuver lab' }));
    expect(screen.getByRole('heading', { name: 'A small what-if.' })).toBeInTheDocument();
    fireEvent.change(screen.getByRole('slider', { name: /delta-v/i }), { target: { value: '0.35' } });
    await user.click(screen.getByRole('button', { name: /preview scenario/i }));
    expect(screen.getByText('LOCAL PREVIEW UPDATED')).toBeInTheDocument();
    expect(screen.getByRole('region', { name: /orbital scene preview/i })).toHaveAttribute('data-path-count', '5');
    expect(screen.getByRole('img', { name: /linearized what-if overlay/i })).toBeInTheDocument();
    expect(screen.getByText(/does not model orbital dynamics/)).toBeInTheDocument();
  });
});
