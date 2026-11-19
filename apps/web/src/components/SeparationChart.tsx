interface SeparationChartProps {
  current: number;
  minimumMeters: number;
  samples: readonly number[];
  replayHalfWindowMinutes?: number;
  comparisonSamples?: readonly number[];
  collisionHalfWindowMinutes?: number;
}

export default function SeparationChart({
  current,
  minimumMeters,
  samples,
  comparisonSamples,
  replayHalfWindowMinutes,
  collisionHalfWindowMinutes,
}: SeparationChartProps) {
  const collisionDemo = collisionHalfWindowMinutes !== undefined;
  if (samples.length < 2) throw new Error('Separation chart requires at least two samples');
  if (comparisonSamples && comparisonSamples.length !== samples.length) {
    throw new Error('Separation chart comparison samples must align with the baseline samples');
  }

  const shownSamples = collisionDemo ? samples.slice(0, Math.floor(samples.length / 2) + 1) : samples;
  const minimum = Math.min(...shownSamples);
  const maximum = Math.max(...shownSamples, ...(comparisonSamples ?? []));
  const chartMinimum = Math.min(minimum, ...(comparisonSamples ?? []));
  const range = Math.max(maximum - chartMinimum, 1);
  const createPath = (values: readonly number[]) => values.map((sample, index) => ({
    x: index * (100 / (samples.length - 1)),
    y: 50 - ((sample - chartMinimum) / range) * 42,
  }));
  const chartPoints = createPath(shownSamples);
  const path = chartPoints.map(({ x, y }, index) => `${index === 0 ? 'M' : 'L'} ${x} ${y}`).join(' ');
  const comparisonPath = comparisonSamples
    ? createPath(comparisonSamples).map(({ x, y }, index) => `${index === 0 ? 'M' : 'L'} ${x} ${y}`).join(' ')
    : undefined;
  const currentX = current;
  const halfWindow = collisionHalfWindowMinutes ?? replayHalfWindowMinutes ?? 30;
  const minimumPoint = chartPoints[replayHalfWindowMinutes ? Math.floor(samples.length / 2) : samples.indexOf(minimum)];
  if (!minimumPoint) throw new Error('Separation chart has no minimum sample');

  return (
    <div className="chart-frame">
      <svg
        className="separation-chart"
        viewBox="0 0 100 56"
        preserveAspectRatio="none"
        role="img"
        aria-label={collisionDemo ? 'Scripted collision approach with zero separation at impact' : replayHalfWindowMinutes ? `SGP4 separation profile with a screened miss distance of ${minimumMeters.toLocaleString()} meters${comparisonPath ? ' and a two-body what-if overlay' : ''}` : `Synthetic separation profile with a fixture minimum of ${minimumMeters.toLocaleString()} meters${comparisonPath ? ' and a linearized what-if overlay' : ''}`}
      >
        <defs>
          <linearGradient id="chart-fill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="var(--selected)" stopOpacity=".16" />
            <stop offset="100%" stopColor="var(--selected)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={`${path} L ${collisionDemo ? 50 : 100} 56 L 0 56 Z`} fill="url(#chart-fill)" />
        {[14, 28, 42].map((y) => <line key={y} x1="0" x2="100" y1={y} y2={y} className="chart-grid" />)}
        <path d={path} className="chart-line" />
        {comparisonPath && <path d={comparisonPath} className="chart-hypothetical-line" />}
        <line x1={currentX} x2={currentX} y1="0" y2="56" className="chart-current" />
        <circle cx={minimumPoint.x} cy={minimumPoint.y} r="2.1" className="chart-tca-dot" />
      </svg>
      <div className="chart-axis"><span>T−{halfWindow}m</span><span>{collisionDemo ? 'Impact' : 'Closest approach'}</span><span>T+{halfWindow}m</span></div>
      {comparisonPath && (
        <div className="chart-legend">
          <span><i className="legend-dot original" />Baseline</span>
          <span><i className="legend-dot hypothetical" />What-if</span>
        </div>
      )}
      <details className="chart-data">
        <summary>View chart data</summary>
        <table>
          <caption>{replayHalfWindowMinutes ? 'SGP4 replay samples' : 'Synthetic fixture path samples'}</caption>
          <thead><tr><th scope="col">Time offset</th><th scope="col">Baseline</th>{comparisonSamples && <th scope="col">What-if</th>}</tr></thead>
          <tbody>
          {shownSamples.map((sample, index) => (
            <tr key={index}>
              <td>{Number((-halfWindow + index * (2 * halfWindow / (samples.length - 1))).toFixed(replayHalfWindowMinutes ? 2 : 0))} min</td>
              <td>{Math.round(sample).toLocaleString()} m</td>
              {comparisonSamples && <td>{Math.round(comparisonSamples[index]).toLocaleString()} m</td>}
            </tr>
            ))}
          </tbody>
        </table>
      </details>
      <p className="chart-note">{collisionDemo ? 'Scripted separation through impact' : replayHalfWindowMinutes ? (comparisonPath ? 'SGP4 baseline and two-body what-if in TEME' : 'SGP4 samples in TEME') : comparisonPath ? 'Linearized fixture preview · not propagation' : 'Simplified synthetic paths · not operational data'}</p>
    </div>
  );
}
