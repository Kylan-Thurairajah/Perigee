import { useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';
import type { SceneSelection } from './types';

interface EncounterMarkerProps {
  position: [number, number, number];
  selected: boolean;
  onSelect: (selection: SceneSelection) => void;
  encounterId: string;
  reducedMotion: boolean;
  paused: boolean;
}

export function EncounterMarker({
  position,
  selected,
  onSelect,
  encounterId,
  reducedMotion,
  paused,
}: EncounterMarkerProps) {
  const core = useRef<THREE.Mesh>(null);
  const haloMaterial = useMemo(
    () => new THREE.MeshBasicMaterial({ color: '#eba16f', transparent: true, opacity: 0.1 }),
    [],
  );
  const halo = useRef<THREE.Mesh>(null);
  const material = useMemo(
    () => new THREE.MeshStandardMaterial({
      color: '#eba16f',
      emissive: '#eba16f',
      emissiveIntensity: selected ? 0.22 : 0.08,
      roughness: 0.7,
    }),
    [selected],
  );

  useFrame((state) => {
    if (reducedMotion || paused) return;
    const pulse = (Math.sin(state.clock.elapsedTime * 2.2) + 1) / 2;
    if (core.current) core.current.scale.setScalar(0.94 + pulse * 0.18);
    if (halo.current) halo.current.scale.setScalar(1.45 + pulse * 0.8);
    haloMaterial.opacity = 0.07 + pulse * 0.14;
  });

  return (
    <group
      position={position}
      rotation={[0, 0, Math.PI / 4]}
      onClick={(event) => {
        event.stopPropagation();
        onSelect({ kind: 'encounter', id: encounterId });
      }}
    >
      <mesh ref={core} material={material} scale={selected ? 1.25 : 1}>
        <boxGeometry args={[0.18, 0.18, 0.18]} />
      </mesh>
      <mesh ref={halo} material={haloMaterial} scale={1.45}>
        <sphereGeometry args={[0.12, 12, 12]} />
      </mesh>
    </group>
  );
}
