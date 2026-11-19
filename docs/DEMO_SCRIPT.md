# Perigee Video Demo

**Status:** walkthrough is ready; video/GIF capture is still pending.

## Before recording

1. Start the app from the repository root with `docker compose up --build -d`.
2. Open `http://localhost:5173` in a browser at 1920×1080 (or 1440×900).
3. Use light theme, select Explorer, and reset the Maneuver Lab if a previous preview is active.
4. Record the browser content without showing credentials, unrelated windows, or unsupported benchmark claims.

## 45-second walkthrough

| Time | Action | Narration / on-screen point |
|---|---|---|
| 0–5s | Hold on Explorer and the globe. | “Perigee is a local-first way to explore orbital traffic.” The objects and orbit tracks are deterministic fictional fixtures. |
| 5–11s | Select the featured encounter. | Show Peregrine A × Lyra B and the 742 m fixture miss distance. Say that separation is not collision probability. |
| 11–22s | Play the timeline briefly, pause, then return the thumb to the TCA marker. | Show the scene, UTC time, and separation profile moving together. Hold on the centered closest-approach marker. |
| 22–36s | Open Maneuver Lab, leave or set Δv to +0.50 m/s, then click **Preview scenario**. | Show the dashed what-if path and baseline/preview profile. The displayed comparison is a linearized fixture illustration, not orbit propagation. |
| 36–45s | Hold on the changed TCA value and the limitation note. | Close on the local, synthetic-demo caveat. Do not describe the preview as a real maneuver recommendation. |

## Capture notes

- For a short impact clip, open **Encounter** and select **Play fictional collision**. At 1× the scripted hit occurs four seconds into an eight-second loop. Use **Pause** to hold the burst or **Restart collision replay** to return to the approach. The fictional impact and fragments are illustrative effects.

- Keep the recording between 35 and 60 seconds; use a steady cursor and pause briefly after each transition.
- Capture at 1080p or higher with browser chrome hidden if practical. Use a screen recorder available on your machine; no recorder appeared in the apps available to this session. Export the video before making any GIF.
- A short silent loop may be cut from the Explorer-to-Encounter transition after reviewing the full video.
- Intended media locations, once files have actually been captured and reviewed: `media/perigee-demo.mp4` and `media/perigee-preview.gif`.
- Do not add a placeholder video, synthetic terminal benchmark, or unverified performance statistic.

## Verified demo boundary

The 45-second walkthrough uses synthetic catalog, encounter, and Maneuver Lab fixtures. The fixture preview applies a linear along-track offset to sampled positions.

## Computed screening walkthrough

Use [screening-demo.omm.json](../data/reference/screening-demo.omm.json) directly. The first object is an OMM transcription of the historical Vanguard verification case from Vallado’s [SGP4-VER suite](https://celestrak.org/publications/AIAA/2006-6753/). The second object is a synthetic variant with a changed catalog ID, mean motion, and mean anomaly. Enter `Vallado verification fixture` as the source name.

1. In **Screening**, import the JSON array. Set retrieval time to the local equivalent of 2000-06-28 00:00 UTC and maximum element age to seven days.
2. Enter the local equivalents of 2000-06-27 19:07 to 19:13 UTC, a 10,000 m threshold, and a 60-second coarse step. Run the screen.
3. Open the result's replay. Move the scrubber so the scene, sampled separation, and chart advance together.
4. Set the educational delta-v to +0.5 m/s and select **Preview what-if**. Show the separation chart overlay and the two minimums across the replay samples. Explain that the impulse is fixed to the replay start and the approximation uses central Earth gravity to estimate the incremental response.
5. Clear the what-if to return to the SGP4 baseline. Keep the model caveat visible in the capture.
