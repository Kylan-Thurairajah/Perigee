import { useLayoutEffect, useMemo, useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import * as THREE from 'three';

interface EarthProps {
  palette: 'light' | 'dark';
  reducedMotion: boolean;
  paused: boolean;
}

interface SurfaceVoxel {
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

function isCloud(longitude: number, latitude: number): boolean {
  if (Math.abs(latitude) > 1.15) return false;
  const band = Math.sin(longitude * 2.1 + latitude * 3.4 + Math.sin(latitude * 5) * 0.45);
  const wisps = Math.sin(longitude * 7.3 - latitude * 2.8);
  return band > 0.78 && wisps > -0.1;
}

function createSurfaceVoxels(): SurfaceVoxel[] {
  const voxels: SurfaceVoxel[] = [];
  const rows = 28;
  const columns = 56;
  const radius = 2.035;

  for (let row = 1; row < rows - 1; row += 1) {
    const latitude = Math.PI / 2 - ((row + 0.5) / rows) * Math.PI;
    for (let column = 0; column < columns; column += 1) {
      const longitude = ((column + 0.5) / columns) * Math.PI * 2 - Math.PI;
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
      const land = isLand(longitude, latitude);
      const cloud = !land && isCloud(longitude, latitude);
      const polarIce = !land && Math.abs(latitude) > 1.3;
      const color = land
        ? new THREE.Color().setHSL(0.34 + variation * 0.035, 0.38 + variation * 0.12, 0.46 + variation * 0.12)
        : cloud || polarIce
          ? new THREE.Color().setHSL(0.52, 0.17, 0.83 + variation * 0.08)
          : new THREE.Color().setHSL(0.56 + variation * 0.025, 0.42 + variation * 0.08, 0.42 + variation * 0.12);
      const scale = cloud ? 0.055 + variation * 0.018 : 0.105 + variation * 0.015;
      voxels.push({ position, quaternion, color, scale });
    }
  }
  return voxels;
}

export function Earth({ palette, reducedMotion, paused }: EarthProps) {
  const group = useRef<THREE.Group>(null);
  const surfaceMesh = useRef<THREE.InstancedMesh>(null);
  const atmosphereMesh = useRef<THREE.Mesh>(null);
  const atmosphereMaterial = useRef<THREE.MeshBasicMaterial>(null);
  const surfaceVoxels = useMemo(createSurfaceVoxels, []);
  const ocean = palette === 'dark' ? '#3d7699' : '#5d94bb';
  const atmosphere = palette === 'dark' ? '#263f53' : '#aecbda';

  useLayoutEffect(() => {
    const mesh = surfaceMesh.current;
    if (!mesh) return;
    const matrix = new THREE.Matrix4();
    for (let index = 0; index < surfaceVoxels.length; index += 1) {
      const voxel = surfaceVoxels[index];
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
  }, [surfaceVoxels]);

  useFrame((state, delta) => {
    if (reducedMotion || paused) return;
    if (group.current) group.current.rotation.y += delta * 0.018;
    const pulse = (Math.sin(state.clock.elapsedTime * 0.65) + 1) / 2;
    atmosphereMesh.current?.scale.setScalar(1.025 + pulse * 0.006);
    if (atmosphereMaterial.current) atmosphereMaterial.current.opacity = 0.045 + pulse * 0.025;
  });

  return (
    <group ref={group}>
      <mesh>
        <icosahedronGeometry args={[2, 3]} />
        <meshStandardMaterial color={ocean} roughness={0.96} flatShading />
      </mesh>
      <instancedMesh ref={surfaceMesh} args={[undefined, undefined, surfaceVoxels.length]} castShadow receiveShadow>
        <boxGeometry args={[1, 1, 1]} />
        <meshStandardMaterial roughness={0.9} flatShading />
      </instancedMesh>
      <mesh ref={atmosphereMesh} scale={1.025}>
        <icosahedronGeometry args={[2, 3]} />
        <meshBasicMaterial ref={atmosphereMaterial} color={atmosphere} transparent opacity={0.045} side={THREE.BackSide} />
      </mesh>
    </group>
  );
}
