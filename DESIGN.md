---
name: HOMIN-DEV Ingress
description: Personal ingress — WebGL Go gophers with links to projects and portfolio sites
colors:
  bg: "#1a1a1a"
  surface: "#242424"
  ink: "#f2f2f2"
  ink-muted: "#b8b8b8"
  ink-faint: "#8a8a8a"
  accent: "#00add8"
  accent-soft: "#00aaaa"
  border: "#ffffff29"
  link-hover: "#00add824"
  shade: "#0000008c"
typography:
  display:
    fontFamily: "Bricolage Grotesque, Segoe UI, sans-serif"
    fontSize: "clamp(1.5rem, 2.5vw + 1rem, 2.25rem)"
    fontWeight: 800
    lineHeight: 1.2
    letterSpacing: "-0.02em"
  title:
    fontFamily: "Bricolage Grotesque, Segoe UI, sans-serif"
    fontSize: "clamp(1.05rem, 0.4vw + 1rem, 1.25rem)"
    fontWeight: 700
    lineHeight: 1.2
    letterSpacing: "normal"
  body:
    fontFamily: "Bricolage Grotesque, Segoe UI, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  label:
    fontFamily: "Bricolage Grotesque, Segoe UI, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
  caption:
    fontFamily: "Bricolage Grotesque, Segoe UI, sans-serif"
    fontSize: "0.8125rem"
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: "normal"
rounded:
  sm: "6px"
  md: "10px"
spacing:
  xs: "4px"
  sm: "8px"
  md: "16px"
  lg: "24px"
  xl: "32px"
components:
  site-header:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "display"
    rounded: "{rounded.sm}"
    padding: "8px 16px"
  link-card:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink}"
    typography: "title"
    rounded: "{rounded.md}"
    padding: "12px 20px"
    height: "44px"
  link-card-hover:
    backgroundColor: "{colors.link-hover}"
    textColor: "{colors.ink}"
  back-link:
    backgroundColor: "{colors.surface}"
    textColor: "{colors.ink-muted}"
    typography: "label"
    rounded: "{rounded.sm}"
    padding: "8px 16px"
    height: "40px"
  fish-quote:
    backgroundColor: "{colors.bg}"
    textColor: "{colors.ink}"
    typography: "title"
    rounded: "{rounded.sm}"
    padding: "16px 24px"
    width: "42ch"
---

## Overview

HOMIN-DEV ingress is a dark, single-viewport brand surface: a WebGL field of Go gophers is the hero, and a quiet typographic chrome (wordmark, destination links, optional movie quote, footer) sits on top. Color strategy is **restrained** — near-black neutrals with Go cyan (`#00ADD8`) as the sole accent. Motion belongs to the 3D scene; UI motion is limited to short ease-out color transitions and respects `prefers-reduced-motion`.

## Colors

- **bg / surface** (`#1a1a1a`, `#242424`): scene and chrome base; keep the canvas and page background aligned.
- **ink ramp** (`#f2f2f2` → `#b8b8b8` → `#8a8a8a`): solid hex muted colors (not translucent white) so contrast stays stable over the animated scene.
- **accent** (`#00add8` / `#00aaaa`): Go cyan for focus rings, hover borders, and grid glow — never as a full-bleed wash.
- **shade** (`rgba(0,0,0,0.55)`): wordmark text-shadow only, for legibility over the WebGL field.
- **Do not** introduce glassmorphism, purple gradients, or a second saturated brand hue.
- **Typography note:** One family (Bricolage Grotesque) with weight contrast is intentional for this surface; do not add a second face unless the brand expands.

## Typography

Single family: **Bricolage Grotesque** (400 / 500 / 700 / 800). Display uses fluid `clamp()` and slight negative tracking (`-0.02em`, never tighter than `-0.04em`). Title/body/label/caption follow a clear step down; captions stay ≥0.8125rem for AA-friendly size on dark UI. Quotes may use italic weight 500; UI chrome stays upright.

## Elevation

Flat tonal layering only. No drop shadows on chrome. Depth comes from the WebGL scene (fog, grid glow) and light translucent surfaces via `color-mix` against `--color-bg` / `--color-surface`. Focus uses a dual-ring (`bg` + `accent`), not glow shadows.

## Components

- **site-header**: text link wordmark, centered, no pill/glass shell; hover shifts to accent.
- **back-link**: compact bordered button (`← Home`) rendered above `.links-container` on non-root paths (`/social`, `/support`, etc.) for explicit return to `/`; cyan arrow (`--color-accent`), hover/active tint matching `.link-card`.
- **link-card**: bordered destination control, min height 44px; hover/active tint with accent, `cursor: pointer`, no lift/shadow.
- **about-modal**: flat tonal dialog (`--color-surface` over `--color-shade`), circular hero image (`border-radius: 50%`, cyan border), prominent name heading with subtitle/title underneath, and Markdown description; opened via `"👱🏻‍♂️ About"` button on `/` or directly at `/about`.
- **wip-modal**: dialog presenting WIP graphic (`wip.webp`) over `--color-shade`, centered title, and close control; opened via `/wip`.
- **fish-quote**: optional blockquote, max ~42ch, single border (not a side stripe).
- **footer**: caption ink-faint, centered, always present (including gophers-only mode); respect `safe-area-inset-bottom`.

## Do's and Don'ts

**Do**
- Keep the gopher scene as the primary visual identity.
- Use CSS variables in `frontend/src/style.css` for any new chrome color or type size.
- Prefer pointer events over mouse-only APIs.
- Cap WebGL `devicePixelRatio` at 2.
- Use `100dvh` and `env(safe-area-inset-*)` for chrome placement on notched devices.

**Don't**
- Restore Inter, glass blur panels, or card lift shadows.
- Use gradient text, side-stripe accents, or emoji in the wordmark.
- Animate layout properties or ignore `prefers-reduced-motion`.
- Inject API strings with `innerHTML`.
