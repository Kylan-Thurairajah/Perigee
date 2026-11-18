import { useEffect, useMemo, useRef } from 'react';
import type { ComponentRef } from 'react';
import { OrbitControls } from '@react-three/drei';
import { useFrame, useThree } from '@react-three/fiber';
import * as THREE from 'three';
import type { SceneMode, SceneSelection } from './types';

interface CameraRigProps {
  mode: SceneMode;
  selected: SceneSelection;
  focusPositions: ReadonlyMap<string, [number, number, number]>;
  reducedMotion: boolean;
}

export function CameraRig({ mode, selected, focusPositions, reducedMotion }: CameraRigProps) {
  const controls = useRef<ComponentRef<typeof OrbitControls>>(null);
  const { camera } = useThree();
  const focusKey = selected ? `${selected.kind}:${selected.id}` : 'overview';
  const moving = useRef(false);
  const following = useRef(false);
  const selectedRef = useRef(selected);
  const focusPositionsRef = useRef(focusPositions);
  selectedRef.current = selected;
  focusPositionsRef.current = focusPositions;
  const desiredTarget = useRef(new THREE.Vector3());
  const desiredPosition = useRef(new THREE.Vector3(8, 6, 9));
  const targetOffset = useMemo(
    () => new THREE.Vector3(7, 4.5, 7).normalize(),
    [],
  );

  useEffect(() => {
    const currentSelection = selectedRef.current;
    const focus = currentSelection ? focusPositionsRef.current.get(currentSelection.id) : undefined;
    desiredTarget.current.set(...(focus ?? [0, 0, 0]));
    const isFocused = Boolean(currentSelection);
    const distance = mode === 'explore' ? (isFocused ? 6.4 : 9.5) : 5.2;
    desiredPosition.current.copy(desiredTarget.current).addScaledVector(targetOffset, distance);
    moving.current = !reducedMotion;
    following.current = Boolean(currentSelection && mode !== 'explore' && !reducedMotion);
    if (reducedMotion) {
      camera.position.copy(desiredPosition.current);
      camera.lookAt(desiredTarget.current);
      if (controls.current) controls.current.target.copy(desiredTarget.current);
    }
  }, [camera, focusKey, mode, reducedMotion, targetOffset]);

  useFrame((_, delta) => {
    if ((!moving.current && !following.current) || !controls.current) return;
    const currentSelection = selectedRef.current;
    if (following.current && currentSelection) {
      const currentFocus = focusPositionsRef.current.get(currentSelection.id);
      if (currentFocus) {
        desiredTarget.current.set(...currentFocus);
        desiredPosition.current.copy(desiredTarget.current).addScaledVector(targetOffset, 5.2);
      }
    }
    const amount = 1 - Math.exp(-delta * 2.8);
    camera.position.lerp(desiredPosition.current, amount);
    controls.current.target.lerp(desiredTarget.current, amount);
    controls.current.update();
    if (
      camera.position.distanceToSquared(desiredPosition.current) < 0.001 &&
      controls.current.target.distanceToSquared(desiredTarget.current) < 0.001
    ) {
      moving.current = false;
    }
  });

  return (
    <OrbitControls
      ref={controls}
      enableDamping
      dampingFactor={0.08}
      minDistance={4.2}
      maxDistance={18}
      maxPolarAngle={Math.PI * 0.88}
      onStart={() => {
        moving.current = false;
        following.current = false;
      }}
    />
  );
}
