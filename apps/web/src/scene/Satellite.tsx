import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { SceneSelection, SatelliteVisual } from './types';

interface SatelliteProps {
  satellite: SatelliteVisual;
  position: readonly [number, number, number];
  selected: boolean;
  dimmed: boolean;
  palette: 'light' | 'dark';
  reducedMotion: boolean;
  paused: boolean;
  onSelect: (selection: SceneSelection) => void;
}

export function Satellite({
  satellite,
  position,
  selected,
  dimmed,
  palette,
  reducedMotion,
  paused,
  onSelect,
}: SatelliteProps) {
  const body = useRef<THREE.Group>(null);
  const panels = useRef<THREE.Group>(null);
  const radialOrientation = useMemo(
    () => new THREE.Quaternion().setFromUnitVectors(
      new THREE.Vector3(0, 1, 0),
      new THREE.Vector3(position[0], position[1], position[2]).normalize(),
    ),
    [position],
  );
  const panelMaterial = useMemo(
    () => new THREE.MeshStandardMaterial({
      color: palette === 'dark' ? '#476e8d' : '#365c7b',
      metalness: 0.2,
      roughness: 0.48,
      transparent: dimmed,
      opacity: dimmed ? 0.38 : 1,
    }),
    [dimmed, palette],
  );
  const panelCellMaterial = useMemo(
    () => new THREE.MeshStandardMaterial({
      color: palette === 'dark' ? '#a9c5d8' : '#91b5ce',
      metalness: 0.16,
      roughness: 0.42,
      transparent: dimmed,
      opacity: dimmed ? 0.38 : 1,
    }),
    [dimmed, palette],
  );
  const hardwareMaterial = useMemo(
    () => new THREE.MeshStandardMaterial({
      color: palette === 'dark' ? '#d4c8a9' : '#a7926f',
      metalness: 0.38,
      roughness: 0.42,
      transparent: dimmed,
      opacity: dimmed ? 0.38 : 1,
    }),
    [dimmed, palette],
  );
  const bodyColor = satellite.color ?? (palette === 'dark' ? '#f2eee4' : '#faf8f1');

  useFrame((state, delta) => {
    if (reducedMotion || paused) return;
    if (body.current) body.current.rotation.y += delta * 0.3;
    if (panels.current) {
      panels.current.rotation.z = Math.sin(state.clock.elapsedTime * 0.85 + satellite.id.length) * 0.08;
    }
  });

  return (
    <group
      position={[position[0], position[1], position[2]]}
      quaternion={radialOrientation}
      scale={selected ? 1.12 : 1}
      onClick={(event) => {
        event.stopPropagation();
        onSelect({ kind: 'satellite', id: satellite.id });
      }}
    >
      <group ref={body}>
        <mesh>
          <boxGeometry args={[0.18, 0.18, 0.18]} />
          <meshStandardMaterial
            color={bodyColor}
            metalness={0.28}
            roughness={0.48}
            transparent={dimmed}
            opacity={dimmed ? 0.38 : 1}
          />
        </mesh>
        <mesh position={[0, 0.15, 0]} material={hardwareMaterial}>
          <cylinderGeometry args={[0.012, 0.026, 0.12, 6]} />
        </mesh>
        <mesh position={[0, 0.22, 0]} material={hardwareMaterial}>
          <coneGeometry args={[0.035, 0.045, 6]} />
        </mesh>
      </group>
      <group ref={panels}>
        {[-1, 1].map((side) => (
          <group key={side} position={[side * 0.31, 0, 0]}>
            <mesh material={panelMaterial}>
              <boxGeometry args={[0.27, 0.045, 0.2]} />
            </mesh>
            <mesh position={[0, 0.024, 0]} material={panelCellMaterial}>
              <boxGeometry args={[0.23, 0.006, 0.16]} />
            </mesh>
            {[-0.045, 0.045].map((offset) => (
              <mesh key={`row-${offset}`} position={[0, 0.028, offset]} material={panelMaterial}>
                <boxGeometry args={[0.23, 0.005, 0.006]} />
              </mesh>
            ))}
            {[-0.055, 0.055].map((offset) => (
              <mesh key={`column-${offset}`} position={[offset, 0.028, 0]} material={panelMaterial}>
                <boxGeometry args={[0.006, 0.005, 0.16]} />
              </mesh>
            ))}
          </group>
        ))}
        <mesh position={[0, 0, 0]} material={hardwareMaterial}>
          <boxGeometry args={[0.16, 0.06, 0.08]} />
        </mesh>
      </group>
      {selected && (
        <mesh>
          <sphereGeometry args={[0.52, 16, 16]} />
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
