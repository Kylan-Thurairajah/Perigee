const fragments = Array.from({ length: 14 }, (_, index) => {
  const z = (index + 0.5) / 7 - 1;
  const radius = Math.sqrt(1 - z * z);
  const angle = index * Math.PI * (3 - Math.sqrt(5));
  return {
    direction: [radius * Math.cos(angle), radius * Math.sin(angle), z],
    distance: 0.33 + (index % 5) * 0.09,
    size: 0.029 + (index % 3) * 0.012,
  };
});

export default function CollisionBurst({ position, progress, reducedMotion }: {
  position: [number, number, number];
  progress: number;
  reducedMotion: boolean;
}) {
  const spread = reducedMotion ? 0.45 : 1 - (1 - progress) ** 2;
  const opacity = reducedMotion ? 0.7 : (1 - progress) ** 1.5;
  return (
    <group position={position}>
      <mesh scale={0.12 + spread * 0.43}>
        <sphereGeometry args={[1, 18, 12]} />
        <meshBasicMaterial color="#ffae62" wireframe transparent opacity={opacity * 0.6} depthWrite={false} />
      </mesh>
      <mesh scale={0.1 * (1 - spread) + 0.02}>
        <sphereGeometry args={[1, 12, 8]} />
        <meshBasicMaterial color="#fff1bc" transparent opacity={opacity} depthWrite={false} />
      </mesh>
      {fragments.map(({ direction, distance, size }, index) => (
        <mesh
          key={index}
          position={[direction[0] * distance * spread, direction[1] * distance * spread, direction[2] * distance * spread]}
          scale={size}
        >
          <boxGeometry />
          <meshBasicMaterial color={index % 2 ? '#e76b36' : '#ffbf6e'} transparent opacity={opacity} depthWrite={false} />
        </mesh>
      ))}
    </group>
  );
}
