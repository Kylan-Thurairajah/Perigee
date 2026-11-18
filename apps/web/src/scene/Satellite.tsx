import { useMemo } from 'react';
import * as THREE from 'three';
import type { SceneSelection, SatelliteVisual } from './types';

interface SatelliteProps {
  satellite: SatelliteVisual;
  position: readonly [number, number, number];
  selected: boolean;
  dimmed: boolean;
  palette: 'light' | 'dark';
  onSelect: (selection: SceneSelection) => void;
}

export function Satellite({
  satellite,
  position,
  selected,
  dimmed,
  palette,
  onSelect,
}: SatelliteProps) {
  const panelMaterial = useMemo(
    () => new THREE.MeshStandardMaterial({ color: palette === 'dark' ? '#9eb9cc' : '#829fb9', metalness: 0.3, roughness: 0.55 }),
    [palette],
  );
  const bodyColor = satellite.color ?? (palette === 'dark' ? '#f2eee4' : '#faf8f1');

  return (
    <group
      position={[position[0], position[1], position[2]]}
      scale={selected ? 1.12 : 1}
      onClick={(event) => {
        event.stopPropagation();
        onSelect({ kind: 'satellite', id: satellite.id });
      }}
    >
      <mesh>
        <boxGeometry args={[0.16, 0.16, 0.16]} />
        <meshStandardMaterial color={bodyColor} metalness={0.28} roughness={0.48} transparent opacity={dimmed ? 0.38 : 1} />
      </mesh>
      <mesh position={[-0.22, 0, 0]} material={panelMaterial}>
        <boxGeometry args={[0.27, 0.055, 0.17]} />
      </mesh>
      <mesh position={[0.22, 0, 0]} material={panelMaterial}>
        <boxGeometry args={[0.27, 0.055, 0.17]} />
      </mesh>
      <mesh position={[0, 0.13, 0]}>
        <cylinderGeometry args={[0.015, 0.025, 0.13, 5]} />
        <meshStandardMaterial color={palette === 'dark' ? '#d4c8a9' : '#a7926f'} />
      </mesh>
      {selected && (
        <mesh>
          <sphereGeometry args={[0.29, 12, 12]} />
          <meshBasicMaterial
            color={palette === 'dark' ? '#a8c6df' : '#5f88ad'}
            transparent
            opacity={0.18}
            side={THREE.BackSide}
          />
        </mesh>
      )}
    </group>
  );
}
