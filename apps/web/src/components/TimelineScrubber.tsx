import { useEffect } from 'react';
import type { KeyboardEvent } from 'react';

interface TimelineScrubberProps {
  value: number;
  onChange: (value: number) => void;
  playing: boolean;
  onPlayingChange: (playing: boolean) => void;
  speed: number;
  onSpeedChange: (speed: number) => void;
  tcaLabel: string;
  utcLabel: string;
}

export default function TimelineScrubber({
  value,
  onChange,
  playing,
  onPlayingChange,
  speed,
  onSpeedChange,
  tcaLabel,
  utcLabel,
}: TimelineScrubberProps) {
  useEffect(() => {
    if (!playing) return;
    const timer = window.setInterval(() => {
      onChange((value + 1) % 101);
    }, Math.max(100, 550 / speed));
    return () => window.clearInterval(timer);
  }, [onChange, playing, speed, value]);

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.code === 'Space') {
      event.preventDefault();
      onPlayingChange(!playing);
    }
    if (event.key === 'ArrowLeft') onChange(Math.max(0, value - 1));
    if (event.key === 'ArrowRight') onChange(Math.min(100, value + 1));
  }

  const tcaPosition = 50;
  return (
    <section className="timeline" aria-label="Encounter timeline">
      <div className="timeline-actions">
        <button
          className="play-control"
          aria-label={playing ? 'Pause encounter replay' : 'Play encounter replay'}
          onClick={() => onPlayingChange(!playing)}
        >
          {playing ? <span className="pause-icon" /> : <span className="play-icon" />}
        </button>
        <div className="speed-control" aria-label="Playback speed">
          {[0.5, 1, 2].map((rate) => (
            <button
              key={rate}
              className={speed === rate ? 'speed-active' : ''}
              aria-pressed={speed === rate}
              onClick={() => onSpeedChange(rate)}
            >
              {rate}×
            </button>
          ))}
        </div>
      </div>
      <div className="timeline-track-wrap">
        <div className="timeline-labels">
          <span>T− 30 MIN</span><span className="tca-label">{tcaLabel}</span><span>T+ 30 MIN</span>
        </div>
        <div className="timeline-track">
          <div className="timeline-progress" style={{ width: `${value}%` }} />
          <div className="timeline-tick tick-start" />
          <div className="timeline-tick tick-mid" />
          <div className="timeline-tick tick-end" />
          <div className="tca-marker" style={{ left: `${tcaPosition}%` }} title={`TCA ${tcaLabel}`} />
          <input
            className="timeline-input"
            type="range"
            min="0"
            max="100"
            value={value}
            onChange={(event) => onChange(Number(event.target.value))}
            onKeyDown={handleKeyDown}
            aria-label="Encounter replay time"
            aria-valuetext={`${utcLabel}, ${Math.round((value - tcaPosition) * 0.6)} minutes from closest approach`}
          />
        </div>
      </div>
      <time className="timeline-time" dateTime={utcLabel}>{utcLabel}<span>UTC</span></time>
    </section>
  );
}
