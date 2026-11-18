import { useLayoutEffect, useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

interface EarthProps {
  palette: 'light' | 'dark';
  reducedMotion: boolean;
  paused: boolean;
}

interface LandVoxel {
  position: THREE.Vector3;
  quaternion: THREE.Quaternion;
  color: THREE.Color;
  scale: number;
}

const LAND_REGIONS = [
  [-0.45, 0.55, 0.52, 0.3],
  [-0.25, 0.1, 0.28, 0.45],
  [0.42, 0.4, 0.2, 0.23],
  [0.5, -0.02, 0.16, 0.31],
  [0.83, 0.44, 0.4, 0.27],
  [0.78, -0.35, 0.2, 0.14],
  [-0.82, -0.43, 0.2, 0.12],
];

function isLand(longitude: number, latitude: number): boolean {
  return LAND_REGIONS.some(([centerLon, centerLat, width, height]) => {
    const longitudeDelta = Math.atan2(Math.sin(longitude - centerLon), Math.cos(longitude - centerLon));
    const normalizedX = longitudeDelta / width;
    const normalizedY = (latitude - centerLat) / height;
    const edge = normalizedX * normalizedX + normalizedY * normalizedY;
    const coast = 0.9 + 0.1 * Math.sin(longitude * 27 + latitude * 19) * Math.cos(latitude * 31);
    return edge < coast;
  });
}

function createLandVoxels(): LandVoxel[] {
  const voxels: LandVoxel[] = [];
  const rows = 32;
  const columns = 64;
  const radius = 2.025;

  for (let row = 1; row < rows - 1; row += 1) {
    const latitude = Math.PI / 2 - (row / rows) * Math.PI;
    for (let column = 0; column < columns; column += 1) {
      const longitude = (column / columns) * Math.PI * 2 - Math.PI;
      if (!isLand(longitude, latitude)) continue;
      const normal = new THREE.Vector3(
        Math.cos(latitude) * Math.cos(longitude),
        Math.sin(latitude),
        Math.cos(latitude) * Math.sin(longitude),
      );
      const position = normal.clone().multiplyScalar(radius);
      const quaternion = new THREE.Quaternion().setFromUnitVectors(
        new THREE.Vector3(0, 1, 0),
        normal,
      );
      const variation = Math.sin(column * 12.9898 + row * 78.233) * 0.5 + 0.5;
      const color = new THREE.Color().setHSL(0.36, 0.18 + variation * 0.08, 0.48 + variation * 0.12);
      voxels.push({ position, quaternion, color, scale: 0.07 + variation * 0.025 });
    }
  }
  return voxels;
}

export function Earth({ palette, reducedMotion, paused }: EarthProps) {
  const group = useRef<THREE.Group>(null);
  const landMesh = useRef<THREE.InstancedMesh>(null);
  const landVoxels = useMemo(createLandVoxels, []);
  const ocean = palette === 'dark' ? '#3d7699' : '#5d94bb';
  const atmosphere = palette === 'dark' ? '#263f53' : '#aecbda';

  useLayoutEffect(() => {
    const mesh = landMesh.current;
    if (!mesh) return;
    const matrix = new THREE.Matrix4();
    for (let index = 0; index < landVoxels.length; index += 1) {
      const voxel = landVoxels[index];
      matrix.compose(
        voxel.position,
        voxel.quaternion,
        new THREE.Vector3(voxel.scale, voxel.scale, voxel.scale),
      );
      mesh.setMatrixAt(index, matrix);
      mesh.setColorAt(index, voxel.color);
    }
    mesh.instanceMatrix.needsUpdate = true;
    if (mesh.instanceColor) mesh.instanceColor.needsUpdate = true;
  }, [landVoxels]);

  useFrame((_, delta) => {
    if (!group.current || reducedMotion || paused) return;
    group.current.rotation.y += delta * 0.012;
  });

  return (
    <group ref={group}>
      <mesh>
        <icosahedronGeometry args={[2, 4]} />
        <meshStandardMaterial color={ocean} roughness={0.96} flatShading />
      </mesh>
      <instancedMesh ref={landMesh} args={[undefined, undefined, landVoxels.length]} castShadow receiveShadow>
        <boxGeometry args={[1, 1, 1]} />
        <meshStandardMaterial roughness={0.9} flatShading />
      </instancedMesh>
      <mesh scale={1.025}>
        <icosahedronGeometry args={[2, 3]} />
        <meshBasicMaterial color={atmosphere} transparent opacity={0.055} side={THREE.BackSide} />
      </mesh>
    </group>
  );
}
