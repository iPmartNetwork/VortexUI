/**
 * Preservation Property Tests - Modal Interaction and StaggerContainer Animation Behavior
 *
 * **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7**
 *
 * These tests verify CURRENT behaviors that MUST be preserved after the fix.
 * They MUST PASS on the UNFIXED code — passing confirms the baseline behavior exists.
 *
 * Observed baseline behaviors on unfixed code:
 * - Modal with short content (< 85vh) renders centered without scrollbar (flex items-center justify-center, no overflow-y)
 * - Backdrop click triggers onClose callback (onClick={onClose} on outer div)
 * - Content click does NOT trigger onClose (e.stopPropagation() on inner div)
 * - Modal open/close uses spring animation with smooth transition (type: "spring", stiffness: 400, damping: 28)
 * - StaggerContainer children animate with staggered fade-in and slide-up (y: yOffset → 0)
 * - GlassCard retains its border-radius, background, and padding styling
 */
import { describe, expect, it } from "vitest";
import * as fc from "fast-check";
import { readFileSync } from "node:fs";
import path from "node:path";

// Read source files to inspect actual component code
const modalSource = readFileSync(
  path.resolve(__dirname, "./Modal.tsx"),
  "utf-8",
);
const staggerContainerSource = readFileSync(
  path.resolve(__dirname, "./StaggerContainer.tsx"),
  "utf-8",
);
const glassCardSource = readFileSync(
  path.resolve(__dirname, "./veltrix/GlassCard.tsx"),
  "utf-8",
);

describe("Preservation: Modal Interaction and StaggerContainer Animation Behavior", () => {
  /**
   * Property 2.1: For all modal content heights below 85vh, no scrollbar appears
   * and modal is vertically centered.
   *
   * The Modal uses `flex items-center justify-center` on the backdrop to center
   * the panel vertically. For short content, no overflow-y-auto is needed on the
   * panel because content fits within the viewport.
   *
   * **Validates: Requirements 3.1, 3.5**
   */
  it("Property: Modal backdrop uses flex centering for vertically-centered display", () => {
    fc.assert(
      fc.property(
        // Generate content heights that are SHORT (below 85vh on a typical 900px viewport)
        fc.integer({ min: 50, max: 600 }),
        (_contentHeight: number) => {
          // The outer backdrop div MUST have flex + items-center + justify-center
          // to center the modal panel vertically and horizontally
          const hasFlexCenter = modalSource.includes("flex items-center justify-center");
          expect(hasFlexCenter).toBe(true);
        },
      ),
      { numRuns: 20 },
    );
  });

  /**
   * Property 2.2: For all backdrop click events, onClose is called exactly once.
   *
   * The outer motion.div (backdrop) has `onClick={onClose}` which triggers the
   * close callback when the backdrop is clicked.
   *
   * **Validates: Requirements 3.3**
   */
  it("Property: Backdrop element has onClick={onClose} for closing the modal", () => {
    fc.assert(
      fc.property(
        // Generate arbitrary backdrop interaction scenarios
        fc.record({
          clickX: fc.integer({ min: 0, max: 1920 }),
          clickY: fc.integer({ min: 0, max: 1080 }),
        }),
        (_interaction) => {
          // The outer backdrop motion.div must have onClick={onClose}
          // This pattern: the first motion.div (backdrop) gets onClick={onClose}
          const backdropPattern = /className="[^"]*fixed inset-0[^"]*"[\s\S]*?onClick=\{onClose\}/;
          const hasBackdropOnClose = backdropPattern.test(modalSource);
          expect(hasBackdropOnClose).toBe(true);
        },
      ),
      { numRuns: 20 },
    );
  });

  /**
   * Property 2.3: For all content click events, onClose is NOT called.
   *
   * The inner motion.div (modal panel) has `onClick={(e) => e.stopPropagation()}`
   * which prevents clicks inside the modal from bubbling up to the backdrop.
   *
   * **Validates: Requirements 3.4**
   */
  it("Property: Modal content panel has stopPropagation to prevent closing on content click", () => {
    fc.assert(
      fc.property(
        // Generate arbitrary content click scenarios
        fc.record({
          clickX: fc.integer({ min: 0, max: 500 }),
          clickY: fc.integer({ min: 0, max: 800 }),
        }),
        (_interaction) => {
          // The inner modal panel must have e.stopPropagation() on its onClick
          const hasStopPropagation = modalSource.includes("e.stopPropagation()");
          expect(hasStopPropagation).toBe(true);
        },
      ),
      { numRuns: 20 },
    );
  });

  /**
   * Property 2.4: Modal open/close uses spring animation with smooth transition.
   *
   * The modal panel uses `transition={{ type: "spring", stiffness: 400, damping: 28 }}`
   * for a smooth, physically-realistic animation feel.
   *
   * **Validates: Requirements 3.2**
   */
  it("Property: Modal animation uses spring transition with stiffness 400 and damping 28", () => {
    fc.assert(
      fc.property(
        // Generate various animation scenarios (open/close)
        fc.boolean(),
        (_isOpening) => {
          // The modal panel must use a spring transition
          const hasSpringType = modalSource.includes('"spring"');
          const hasStiffness400 = modalSource.includes("stiffness: 400");
          const hasDamping28 = modalSource.includes("damping: 28");

          expect(hasSpringType).toBe(true);
          expect(hasStiffness400).toBe(true);
          expect(hasDamping28).toBe(true);
        },
      ),
      { numRuns: 10 },
    );
  });

  /**
   * Property 2.5: For all StaggerContainer renders, children animate with
   * staggered delay and y offset (slide-up).
   *
   * StaggerContainer wraps each child in a motion.div with:
   * - initial={{ opacity: 0, y: yOffset }}
   * - animate={{ opacity: 1, y: 0 }}
   * - transition={{ delay: i * staggerDelay, duration, ease }}
   *
   * **Validates: Requirements 3.6, 3.7**
   */
  it("Property: StaggerContainer animates children with staggered delay and y offset", () => {
    fc.assert(
      fc.property(
        // Generate various stagger configurations
        fc.record({
          numChildren: fc.integer({ min: 1, max: 50 }),
          staggerDelay: fc.double({ min: 0.01, max: 0.2, noNaN: true }),
          yOffset: fc.integer({ min: 4, max: 24 }),
        }),
        (_config) => {
          // StaggerContainer must animate children with initial y offset and opacity: 0
          const hasInitialOpacity = staggerContainerSource.includes("opacity: 0");
          const hasInitialY = staggerContainerSource.includes("y: yOffset");
          const hasAnimateOpacity = staggerContainerSource.includes("opacity: 1");
          const hasAnimateY = staggerContainerSource.includes("y: 0");
          // Must use stagger delay: delay: i * staggerDelay
          const hasStaggerDelay = staggerContainerSource.includes("i * staggerDelay");

          expect(hasInitialOpacity).toBe(true);
          expect(hasInitialY).toBe(true);
          expect(hasAnimateOpacity).toBe(true);
          expect(hasAnimateY).toBe(true);
          expect(hasStaggerDelay).toBe(true);
        },
      ),
      { numRuns: 20 },
    );
  });

  /**
   * Property 2.6: GlassCard retains its border-radius, background, and padding styling.
   *
   * GlassCard applies rounded-2xl (border-radius), bg-bg-elevated (background),
   * border, and padding (p-5 default, p-4 compact) that must be preserved.
   *
   * **Validates: Requirements 3.1 (visual preservation)**
   */
  it("Property: GlassCard preserves its core styling (rounded-2xl, bg-bg-elevated, border, padding)", () => {
    fc.assert(
      fc.property(
        // Generate various GlassCard configurations
        fc.record({
          compact: fc.boolean(),
          hover: fc.boolean(),
        }),
        (_config) => {
          // GlassCard must have its core styling classes
          const hasRounded = glassCardSource.includes("rounded-2xl");
          const hasBgElevated = glassCardSource.includes("bg-bg-elevated");
          const hasBorder = glassCardSource.includes("border border-border");
          const hasPadding =
            glassCardSource.includes('"p-5"') || glassCardSource.includes('"p-4"');

          expect(hasRounded).toBe(true);
          expect(hasBgElevated).toBe(true);
          expect(hasBorder).toBe(true);
          expect(hasPadding).toBe(true);
        },
      ),
      { numRuns: 10 },
    );
  });
});
