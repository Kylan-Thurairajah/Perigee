import { useEffect, useRef } from 'react';
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
  halfWindowMinutes?: number;
  intervals?: number;
  continuous?: boolean;
  millisecondsPerStep?: number;
  midpointLabel?: string;
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
  halfWindowMinutes = 30,
  intervals = 100,
  continuous = false,
  millisecondsPerStep = 550,
  midpointLabel = 'TCA',
}: TimelineScrubberProps) {
  const position = value / intervals * 100;
  const cursor = useRef({ position: value, emitted: value });
  useEffect(() => {
    if (value !== cursor.current.emitted) cursor.current = { position: value, emitted: value };
  }, [value]);

  useEffect(() => {
    if (!playing) return;
    let previous = performance.now();
    let frame: number;
    function advance(now: number) {
      const elapsed = Math.max(0, now - previous);
      previous = now;
      if (!document.hidden) {
        const cycle = continuous ? intervals : intervals + 1;
        const next = (cursor.current.position + elapsed * speed / millisecondsPerStep) % cycle;
        const display = continuous ? next : Math.floor(next + 1e-8) % cycle;
        if (display !== cursor.current.emitted) onChange(display);
        cursor.current = { position: next, emitted: display };
      }
      frame = window.requestAnimationFrame(advance);
    }
    const resetTime = () => { previous = performance.now(); };
    frame = window.requestAnimationFrame(advance);
    document.addEventListener('visibilitychange', resetTime);
    return () => {
      window.cancelAnimationFrame(frame);
      document.removeEventListener('visibilitychange', resetTime);
    };
  }, [continuous, intervals, millisecondsPerStep, onChange, playing, speed]);

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.code === 'Space') {
      event.preventDefault();
      onPlayingChange(!playing);
    }
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
          <span>T− {halfWindowMinutes} MIN</span><span className="tca-label">{tcaLabel}</span><span>T+ {halfWindowMinutes} MIN</span>
        </div>
        <div className="timeline-track">
          <div className="timeline-progress" style={{ width: `${position}%` }} />
          <div className="timeline-tick tick-start" />
          <div className="timeline-tick tick-mid" />
          <div className="timeline-tick tick-end" />
          <div className="tca-marker" style={{ left: `${tcaPosition}%` }} title={`${midpointLabel} ${tcaLabel}`} />
          <input
            className="timeline-input"
            type="range"
            min="0"
            max={intervals}
            step={continuous ? 'any' : 1}
            value={value}
            onChange={(event) => {
              const next = Number(event.target.value);
              cursor.current = { position: next, emitted: next };
              onChange(next);
            }}
            onKeyDown={handleKeyDown}
            aria-label="Encounter replay time"
            aria-valuetext={`${utcLabel}, ${Number(((position - tcaPosition) * halfWindowMinutes / 50).toFixed(2))} minutes from ${midpointLabel === 'TCA' ? 'closest approach' : 'impact'}`}
          />
        </div>
      </div>
      <time className="timeline-time" dateTime={utcLabel}>{utcLabel}<span>UTC</span></time>
    </section>
  );
}
