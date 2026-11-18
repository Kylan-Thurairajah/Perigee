# Perigee Design

## Design Intent

Perigee combines an airy Apple-like product interface with a playful voxel diorama. It is a desktop web application, not a phone UI. The interface should feel spacious and deliberate, while charts and metrics remain precise.

## Palette

Light theme:

- Canvas: `#F7F6F2`
- Surface: `#FFFFFF`
- Secondary surface: `#EFEEE9`
- Text primary: `#263747`
- Text muted: `#748597`
- Border: `#DEE3E6`
- Orbit blue: `#829FB9`
- Selected blue: `#5F88AD`
- Warning apricot: `#EBA16F`
- Warning pale background: `#FFF2E8`
- Globe sea: `#5D94BB`
- Globe land: `#8BB39C`
- Clouds: restrained near-white

Dark theme:

- Canvas: `#151F29`
- Surface: `#202C37`
- Secondary surface: `#263543`
- Text primary: `#F4F5F2`
- Text muted: `#9FADBA`
- Border: `#344451`
- Orbit blue: `#86A9C5`
- Warning: `#EBA16F`

The warning color indicates an encounter to investigate, never a verified catastrophic collision. Color must be paired with text and iconography.

## Typography and Material

- Use Inter or a modern system font stack.
- Use restrained weights and tight hierarchy.
- Tabular numerals are used for numeric values and timestamps.
- Controls use softly rounded rectangles, hairline borders, and small restrained shadows.
- Avoid heavy blur, glossy glassmorphism, enterprise mission-control chrome, rainbow colors, and giant dashboard tiles.

## Layout

### Explorer

- Approximately 64px top navigation.
- Main 3D canvas occupies 65-75 percent of usable width.
- Context drawer is approximately 320px and collapsible.
- A compact bottom-aligned time control is preferred.
- No permanent large left sidebar.

### Encounter Analysis

- 3D viewport occupies about 60 percent.
- Analytics column occupies about 40 percent.
- Show encounter identity, three key metric tiles, separation curve, object metadata, and Maneuver Lab action.
- Place the timeline below the main viewport.

### Maneuver Lab

- Original and hypothetical 3D paths occupy the left.
- A compact control and recomputed metric panel occupies the right.
- Retain one maneuver type and a visible limitations summary.

At 1440x900, 1280x800, and 1920x1080, the desktop layouts must remain legible and unclipped. Tablet layouts may collapse the analytics panel while preserving the scene. Mobile receives a useful data view and a small-screen note rather than a dedicated design target.

## 3D Direction

- Build a lightly voxelized Earth with stylized procedural geometry.
- Use reusable block or low-poly satellite geometry.
- Keep orbit paths legible, with an understated idle rotation.
- Slow or stop globe motion during direct interaction.
- Use a clean outline or halo for selection rather than glowing rings.
- Frame encounter pairs on selection and let the rest of the scene recede subtly.
- Respect `prefers-reduced-motion`.

## Timeline Scrubber

Use a bespoke thin precision timeline:

- 2px rounded track.
- Small 11-14px thumb that enlarges on hover or keyboard focus.
- Precise tick markers without dense decorative marks.
- Orange TCA diamond with accessible timestamp tooltip.
- Separate play/pause and 0.5x, 1x, and 2x controls.
- Current UTC and T-/T+ offset at the right.
- Left/Right keys seek and Space toggles playback when focused.
- Screen reader labels explain the time and selected result.

The component is named and coded as a timeline scrubber, not a scrollbar.

## Interaction and States

- Panel transitions use approximately 180-350ms ease-out motion.
- Appropriate camera transitions use approximately 700-1200ms.
- Hovered buttons receive a tiny scale and shadow change only.
- Keyboard focus is clearly visible.
- Loading states are quiet and neutral.
- Empty and error states are designed and include a recovery action where possible.
- All essential controls have labels and keyboard operation.
- Chart values are available in tooltips and accessible tables.

## Theme

Light is the default. A compact sun/moon control toggles theme and persists the choice in localStorage. Dark mode changes surfaces and text while preserving semantic colors.

## Scene API Boundary

The lead and UI/scene owners should agree a typed scene boundary comparable to:

```ts
type SceneSelection =
  | { kind: 'satellite'; id: string }
  | { kind: 'encounter'; id: string }
  | null;

type SceneMode = 'explore' | 'encounter' | 'maneuver';

interface OrbitalSceneProps {
  mode: SceneMode;
  sceneTimeUtc: string;
  selected: SceneSelection;
  satellites: SatelliteVisual[];
  paths: OrbitPathVisual[];
  encounter?: EncounterVisual;
  onSelect: (selection: SceneSelection) => void;
}
```

The concrete interface belongs to the lead-owned contract work and must remain synchronized with UI and scene implementations.
