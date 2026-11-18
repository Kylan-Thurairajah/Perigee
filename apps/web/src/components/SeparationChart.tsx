interface SeparationChartProps {
  current: number;
  minimumMeters: number;
}

const points = [42, 37, 34, 30, 26, 22, 18, 14, 11, 9, 8, 11, 17, 24, 31, 39, 45, 50];

export default function SeparationChart({ current, minimumMeters }: SeparationChartProps) {
  const activeIndex = Math.min(points.length - 1, Math.round((current / 100) * (points.length - 1)));
  const path = points.map((point, index) => `${index === 0 ? 'M' : 'L'} ${index * (100 / (points.length - 1))} ${point}`).join(' ');
  const currentX = activeIndex * (100 / (points.length - 1));

  return (
    <div className="chart-frame">
      <svg
        className="separation-chart"
        viewBox="0 0 100 56"
        preserveAspectRatio="none"
        role="img"
        aria-label={`Separation curve with a fixture minimum of ${minimumMeters.toLocaleString()} meters`}
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
        <circle cx="58.8" cy="8" r="2.1" className="chart-tca-dot" />
      </svg>
      <div className="chart-axis"><span>T−30m</span><span>Closest approach</span><span>T+30m</span></div>
      <details className="chart-data">
        <summary>View chart data</summary>
        <table>
          <caption>Illustrative separation series</caption>
          <thead><tr><th scope="col">Sample</th><th scope="col">Relative separation</th></tr></thead>
          <tbody>
            {points.map((point, index) => (
              <tr key={index}><td>{index + 1}</td><td>{Math.round(minimumMeters + point * 175).toLocaleString()} m</td></tr>
            ))}
          </tbody>
        </table>
      </details>
      <p className="chart-note">Fixture preview · sampled curve is illustrative</p>
    </div>
  );
}
