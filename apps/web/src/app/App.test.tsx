import { cleanup, render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import App from './App';

vi.mock('../scene/OrbitalScene', () => ({
  default: () => <div data-testid="orbital-scene" />,
}));

afterEach(() => {
  cleanup();
  localStorage.clear();
  document.documentElement.removeAttribute('data-theme');
});

describe('Perigee application', () => {
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
    await user.click(screen.getByRole('button', { name: /preview scenario/i }));
    expect(screen.getByText('LOCAL PREVIEW UPDATED')).toBeInTheDocument();
    expect(screen.getByText(/does not model orbital dynamics/)).toBeInTheDocument();
  });
});
