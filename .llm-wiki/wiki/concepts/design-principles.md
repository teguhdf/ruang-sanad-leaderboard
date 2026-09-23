# Design Principles

> **Boilerplate frontend principles.** Read brief → infer design direction → apply anti-slop rules.
> Framework agnostic. Contextual — not all rules are active automatically.

---

## 0. BRIEF INFERENCE

Before writing code, **read the room.** Don't jump to default aesthetics.

### 0.A Signals

1. **Page kind** — landing (SaaS/consumer/agency), portfolio, editorial, dashboard/admin, auth page
2. **Vibe words** — "minimalist", "premium", "playful", "brutalist", "editorial", "dark tech", "Apple-y", "B2B serious"
3. **Reference** — URL, screenshot, product, competitor brand
4. **Audience** — B2B procurement vs consumer vs developer. Audience chooses aesthetic, not your taste
5. **Quiet constraints** — accessibility, public sector, regulated industry. These **override** aesthetic preference

### 0.B Design Read

Before generating, output one line:

> *"Reading this as: \<page kind> for \<audience>, with a \<vibe> language."*

### 0.C Anti-Default Discipline

Don't default to: purple gradient, centered hero on dark mesh, 3 equal feature cards, glassmorphism on all cards, Inter + slate-900. These are LLM defaults. **Reach past them deliberately.**

---

## 1. THE THREE DIALS

After design read, set three dials. All layout/motion/density decisions below are gated by these.

See also: [[vibe-minimalist]], [[vibe-premium-consumer]], [[vibe-playful-experimental]], [[vibe-dark-tech]], [[vibe-brutalist]]

> 📖 **Vibe library is optional.** This is just reference — each project can have its own vibe.
> Want to add a vibe? Create a page `concepts/vibe-<name>` and link from here.
> Want to skip vibe? Go straight to the pre-flight checklist.

| Dial | 1 | 10 | Default |
|------|---|----|---------|
| **DESIGN_VARIANCE** | Perfect Symmetry | Artsy Chaos | **8** |
| **MOTION_INTENSITY** | Static | Cinematic / Physics | **6** |
| **VISUAL_DENSITY** | Airy gallery | Cockpit packed | **4** |

### Dial Inference

| Signal / Brief | VARIANCE | MOTION | DENSITY | Related vibe |
|---------------|----------|--------|---------|-------------|
| Minimalist / clean / editorial / Linear-style | 5-6 | 3-4 | 2-3 | [[vibe-minimalist]] |
| Premium consumer / Apple-y / luxury / brand | 7-8 | 5-7 | 3-4 | [[vibe-premium-consumer]] |
| Playful / wild / Awwwards / experimental / agency | 9-10 | 8-10 | 3-4 | [[vibe-playful-experimental]] |
| Dark tech / devtool / hacker | 4-5 | 2-3 | 5-7 | [[vibe-dark-tech]] |
| Brutalist / raw / anti-design | 8-10 | 2-3 | 4-6 | [[vibe-brutalist]] |
| Trust-first / public-sector / regulated / accessibility | 3-4 | 2-3 | 4-5 | — |
| Landing page / portfolio / marketing site (default) | 7-9 | 6-8 | 3-5 | — |
| Redesign - preserve | match existing | +1 | match existing | — |
| Redesign - overhaul | +2 | +2 | match existing | — |

### Use-Case Presets

| Use case | VARIANCE | MOTION | DENSITY |
|----------|----------|--------|---------|
| Landing (SaaS, mainstream) | 7 | 6 | 4 |
| Landing (Agency / creative) | 9 | 8 | 3 |
| Landing (Premium consumer) | 7 | 6 | 3 |
| Portfolio (Designer / studio) | 8 | 7 | 3 |
| Portfolio (Developer) | 6 | 5 | 4 |
| Editorial / Blog | 6 | 4 | 3 |
| Public-sector service | 3 | 2 | 5 |
| Redesign - preserve | match | match+1 | match |
| Redesign - overhaul | +2 | +2 | match |

---

## 2. ANTI AI-SLOP RULES

### 🎨 Color

| ❌ AI Slop | ✅ Replacement |
|-----------|---------|
| Purple/blue glow + dark mesh hero | Intentional solid gradient, or soft radial gradient |
| Warm beige/cream bg (`#f5f1ea`, `#efeae0`) | Cool neutrals (slate/zinc/stone scale) |
| Brass/ochre/oxblood accent (`#b08947`, `#9a2436`) | High-contrast singular accent (emerald, electric blue, deep rose) |
| Random gradient in every section | One gradient in hero only |
| Pure black `#000` dark mode | Navy-tinted near-black (`#0f172a`, `#070b16`) |
| Opacity stacking 3+ layers | Max 2 layer overlay |
| **Premium-consumer palette ban** — beige+brass+espresso (`#f5f1ea` / `#b08947` / `#1a1714`) | Cold luxury, forest, black+tan, cobalt+cream, terracotta+slate |
| Max 1 accent per page | One accent, consistent from hero to footer |

### 🔤 Typography

| ❌ AI Slop | ✅ Replacement |
|-----------|---------|
| Inter as default font | OK (allowed), but don't re-import via Google Fonts. Geist > Inter for modern feel |
| Fraunces / Instrument Serif | **Banned.** Two favorite LLM fonts |
| Serif for "creative"/"premium" | **Default sans-serif.** Serif only if brand explicitly mentions |
| Mixed-family emphasis (sans headline + serif word) | Italic/bold from the SAME font |
| Headline > 8 words | Max 8 words display; >8 use 2 lines |
| Em-dash (`—`) without spaces | `—` with spaces (`word — word`) |

### 📐 Layout

| ❌ AI Slop | ✅ Replacement |
|-----------|---------|
| Centered hero + CTA | Split screen, left-aligned + asset, asymmetric |
| 3 equal feature cards | Size variation (1 large + 2 small), asymmetric grid |
| Cards-inside-cards-inside-cards | Flat hierarchy, border/divider is enough |
| Left text + right image every section | Variation: full-width, grid, overlap, background-image |
| Zigzag alternation > 2 sections | Break with full-width section, bento, marquee |
| Glassmorphism on all cards | Thin border + solid bg. Glass only for overlay/navbar |
| Eyebrow in **every** section | Max 1 eyebrow per 3 sections |
| Split-header (left headline + right explainer) | Stack vertical. Split only if there's a compositional reason |
| Button text wrap on desktop | Shorten label. Max 3 words for primary CTA |
| Duplicate CTA intent | One label per intent. "Contact" + "Get in touch" = pick one |

### 🧩 Components & State

| ❌ AI Slop | ✅ Replacement |
|-----------|---------|
| Loading = spinner only | Skeletal loader matching final layout shape |
| Empty state = empty | Beautiful empty state + instructions |
| Only "success state" implementation | Always implement: loading, empty, error, success |
| Keyboard/Screen reader ignored | Focus visible, label, role, contrast WCAG AA |
| Placeholder-as-label | Label above input. Placeholder only as example |

### 📸 Images

Landing and portfolio are **visual products.** Text-only pages with fake-screenshot divs are slop.

Priority:

1. **Generate images** — if there's an image-gen tool in the environment, use it
2. **Real photos** — picsum.photos, Unsplash, or brand assets
3. **Last resort** — placeholder `<!-- TODO: hero image -->`, don't fake divs

**Logo wall** must be real SVG. No plain text wordmarks. Source: Simple Icons, or inline SVG monogram.

### 🎬 Motion

| Principle | Rule |
|---------|--------|
| Entry animation | Present, but subtle. `fade` + `translateY` is enough |
| Scroll-triggered | Only for hero and section divider |
| Hover cards | `translateY(-2px)` + shadow, or border highlight |
| Button press | `scale(0.97)` — tactile feedback |
| **Banned** | Infinite auto-scroll carousel, parallax everywhere, confetti |

---

## 3. DESIGN SYSTEM MAP

Pick foundation according to brief. Don't invent CSS for things that have official packages.

| Brief | Official Design System |
|-------|----------------------|
| Microsoft/enterprise/dashboard | Fluent UI |
| Google-ish UI | Material Web (Material 3) |
| IBM-style B2B | Carbon |
| GitHub-style devtool | Primer |
| Public-sector UK | GOV.UK Frontend |
| US public-sector | USWDS |
| Modern accessible React | Radix Themes |
| Tailwind SaaS (default) | shadcn/ui (with customization) |

**One system per project.** Don't mix Fluent + Carbon in one tree.

For pure aesthetics (not official systems):

| Aesthetic | Implementation |
|-----------|-------------|
| Glassmorphism | `backdrop-filter` + border highlight + solid fallback |
| Bento grid | CSS Grid. No library owns this |
| Brutalism | Native CSS, monospace, raw borders |
| Editorial | Serif type, asymmetric grid, whitespace |
| Dark tech | Mono + accent neon, terminal motifs |

---

## 4. LAYOUT DISCIPLINE (Hard Rules)

These rules are **mandatory**. Violating = shipping broken work.

### Hero

- **MUST fit in viewport** — headline max 2 lines, subtext max 20 words, CTA visible without scroll
- Font scale: `text-4xl md:text-5xl lg:text-6xl` default. Don't use `text-7xl` for headline >6 words
- Top padding max `pt-24` — more than that is floating
- Max **4 text elements**: eyebrow (optional) + headline + subtext + CTAs
- **Banned in hero:** tagline below CTA, trust strip, pricing teaser, feature list, avatar row, logo wall
- Logo wall = dedicated section **below hero**

### Navigation

- One line on desktop. If it doesn't fit in `lg`: condense labels, drop secondary items, or hamburger
- Height max 80px, default 64-72px

### Bento Grids

- Cell count = content count. 3 items → 3 cells. **No empty cells**
- Rhythm: composition variation, don't do 6 left-image-right-text in a row
- Need visual variation: at least 2-3 cells have image/gradient/pattern, not text-only

### Section Repetition

- One layout family can appear **max 1 time** per page
- Landing 8 sections must use at least **4 different layout families**
- Zigzag (image+text alternating) max **2 consecutive sections**. The 3rd = fail

### Mobile

- Every multi-column layout must declare a `< 768px` fallback
- Navigation single-line on mobile → hamburger/drawer

---

## 5. PRE-FLIGHT CHECKLIST

Before declaring done, check these:

- [ ] Hero fits viewport? Headline ≤2 lines, subtext ≤20 words?
- [ ] Navigation one line on desktop?
- [ ] Eyebrow count ≤ ceil(sectionCount / 3)?
- [ ] Zigzag alternation ≤ 2 consecutive?
- [ ] Loading + empty + error states present?
- [ ] Button text doesn't wrap?
- [ ] CTA intent not duplicated?
- [ ] Contrast WCAG AA (4.5:1 body, 3:1 large)?
- [ ] One accent color per page?
- [ ] no `#000` for dark mode bg?
- [ ] Real images (generated/photo)? Not fake divs?
- [ ] Focus visible + keyboard navigable?
- [ ] `min-h-[100dvh]` not `h-screen` for hero?
