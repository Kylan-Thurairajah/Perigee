import { useMemo } from 'react';
import * as THREE from 'three';
import type { SceneSelection } from './types';

interface EncounterMarkerProps {
  position: [number, number, number];
  selected: boolean;
  onSelect: (selection: SceneSelection) => void;
  encounterId: string;
}

export function EncounterMarker({
  position,
  selected,
  onSelect,
  encounterId,
}: EncounterMarkerProps) {
  const material = useMemo(
    () => new THREE.MeshStandardMaterial({
      color: '#eba16f',
      emissive: '#eba16f',
      emissiveIntensity: selected ? 0.22 : 0.08,
      roughness: 0.7,
    }),
    [selected],
  );
  return (
    <group
      position={position}
      rotation={[0, 0, Math.PI / 4]}
      onClick={(event) => {
        event.stopPropagation();
        onSelect({ kind: 'encounter', id: encounterId });
      }}
    >
      <mesh material={material} scale={selected ? 1.25 : 1}>
        <boxGeometry args={[0.18, 0.18, 0.18]} />
      </mesh>
      <mesh scale={1.75}>
        <sphereGeometry args={[0.12, 12, 12]} />
        <meshBasicMaterial color="#eba16f" transparent opacity={0.15} />
      </mesh>
    </group>
  );
}
