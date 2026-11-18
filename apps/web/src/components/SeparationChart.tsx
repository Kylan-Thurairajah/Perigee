interface SeparationChartProps {
  current: number;
  minimumMeters: number;
  samples: readonly number[];
}

export default function SeparationChart({ current, minimumMeters, samples }: SeparationChartProps) {
  if (samples.length < 2) throw new Error('Separation chart requires at least two fixture samples');

  const minimum = Math.min(...samples);
  const maximum = Math.max(...samples);
  const range = Math.max(maximum - minimum, 1);
  const chartPoints = samples.map((sample, index) => ({
    x: index * (100 / (samples.length - 1)),
    y: 50 - ((sample - minimum) / range) * 42,
  }));
  const path = chartPoints.map(({ x, y }, index) => `${index === 0 ? 'M' : 'L'} ${x} ${y}`).join(' ');
  const currentX = current;
  const minimumPoint = chartPoints[samples.indexOf(minimum)];
  if (!minimumPoint) throw new Error('Separation chart has no minimum sample');

  return (
    <div className="chart-frame">
      <svg
        className="separation-chart"
        viewBox="0 0 100 56"
        preserveAspectRatio="none"
        role="img"
        aria-label={`Synthetic separation profile with a fixture minimum of ${minimumMeters.toLocaleString()} meters`}
      >
        <defs>
          <linearGradient id="chart-fill" x1="0" x2="0" y1="0" y2="1">
            <stop offset="0%" stopColor="var(--selected)" stopOpacity=".16" />
            <stop offset="100%" stopColor="var(--selected)" stopOpacity="0" />
          </linearGradient>
        </defs>
        <path d={`${path} L 100 56 L 0 56 Z`} fill="url(#chart-fill)" />
        {[14, 28, 42].map((y) => <line key={y} x1="0" x2="100" y1={y} y2={y} className="chart-grid" />)}
        <path d={path} className="chart-line" />
        <line x1={currentX} x2={currentX} y1="0" y2="56" className="chart-current" />
        <circle cx={minimumPoint.x} cy={minimumPoint.y} r="2.1" className="chart-tca-dot" />
      </svg>
      <div className="chart-axis"><span>T−30m</span><span>Closest approach</span><span>T+30m</span></div>
      <details className="chart-data">
        <summary>View chart data</summary>
        <table>
          <caption>Synthetic fixture path samples</caption>
          <thead><tr><th scope="col">Time offset</th><th scope="col">Separation</th></tr></thead>
          <tbody>
          {samples.map((sample, index) => (
            <tr key={index}>
              <td>{Math.round(-30 + index * (60 / (samples.length - 1)))} min</td>
              <td>{Math.round(sample).toLocaleString()} m</td>
            </tr>
            ))}
          </tbody>
        </table>
      </details>
      <p className="chart-note">Simplified synthetic paths · not operational data</p>
    </div>
  );
}
