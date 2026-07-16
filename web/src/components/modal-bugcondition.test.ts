/**
 * Bug Condition Exploration Test - Modal Overflow and Stale Layout from Scale Animation
 *
 * **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**
 *
 * This test asserts the EXPECTED (fixed) behavior against the CURRENT (unfixed) code.
 * It MUST FAIL on unfixed code — failure confirms the bug exists.
 *
 * Counterexamples surfaced:
 * - Modal overflows viewport (no max-height constraint, no overflow-y-auto)
 * - Scale animation produces stale dimensions (initial scale: 0.92)
 * - overflow-hidden clips translateY entrance animation
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
const inboundsSource = readFileSync(
  path.resolve(__dirname, "../pages/Inbounds.tsx"),
  "utf-8",
);

describe("Bug Condition Exploration: Modal Overflow and Stale Layout from Scale Animation", () => {
  /**
   * Property 1.1: Modal with tall content MUST have max-height constraint
   *
   * For any modal content height exceeding 85vh, the Modal component MUST
   * include a `max-h-[85vh]` constraint on the panel element.
   *
   * EXPECTED: This test FAILS on unfixed code (no max-h-[85vh] present)
   */
  it("Property: Modal panel includes max-h-[85vh] constraint for tall content", () => {
    fc.assert(
      fc.property(
        // Generate content heights that exceed 85vh (e.g. > 700px on a 900px viewport)
        fc.integer({ min: 700, max: 3000 }),
        (_contentHeight: number) => {
          // The Modal's inner motion.div panel should have max-h-[85vh]
          // This ensures tall content (like our generated height) is constrained
          const hasMaxHeight = modalSource.includes("max-h-[85vh]");
          expect(hasMaxHeight).toBe(true);
        },
      ),
      { numRuns: 10 },
    );
  });

  /**
   * Property 1.2: Modal content area MUST have overflow-y-auto for scrolling
   *
   * For any modal content height exceeding the constrained viewport area, the
   * content area MUST have `overflow-y-auto` or `overflow-y: auto` to allow scrolling.
   *
   * EXPECTED: This test FAILS on unfixed code (no overflow-y-auto present)
   */
  it("Property: Modal content area has overflow-y-auto for scrollable content", () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 700, max: 3000 }),
        (_contentHeight: number) => {
          // The Modal should have an overflow-y-auto wrapper for the children content area
          const hasOverflowYAuto =
            modalSource.includes("overflow-y-auto") ||
            modalSource.includes("overflow-y: auto");
          expect(hasOverflowYAuto).toBe(true);
        },
      ),
      { numRuns: 10 },
    );
  });

  /**
   * Property 1.3: Modal animation MUST NOT use scale transform
   *
   * The scale: 0.92 initial animation causes the browser to calculate child
   * layout at reduced scale, producing stale dimensions. The Modal animation
   * MUST NOT include a scale property in its initial animation props.
   *
   * EXPECTED: This test FAILS on unfixed code (scale: 0.92 IS present)
   */
  it("Property: Modal animation does NOT use scale transform (no stale layout)", () => {
    fc.assert(
      fc.property(
        // Generate various potential scale values that would cause the bug
        fc.double({ min: 0.5, max: 0.99, noNaN: true }),
        (_scaleValue: number) => {
          // The Modal's initial animation should NOT contain any scale property
          // Regex matches patterns like `scale: 0.92` or `scale:0.92`
          const scalePattern = /scale:\s*[\d.]+/;
          const hasScaleAnimation = scalePattern.test(modalSource);
          expect(hasScaleAnimation).toBe(false);
        },
      ),
      { numRuns: 10 },
    );
  });

  /**
   * Property 1.4: GlassCard in Inbounds MUST use overflow-x-hidden (not overflow-hidden)
   *
   * The GlassCard wrapping the StaggerContainer uses `overflow-hidden` which clips
   * the vertical translateY entrance animation. It MUST use `overflow-x-hidden` to
   * only hide horizontal overflow while allowing vertical animation.
   *
   * EXPECTED: This test FAILS on unfixed code (overflow-hidden IS present instead of overflow-x-hidden)
   */
  it("Property: GlassCard in Inbounds uses overflow-x-hidden (not overflow-hidden) to avoid clipping translateY animation", () => {
    fc.assert(
      fc.property(
        // Generate various y-offset values that StaggerContainer might use
        fc.integer({ min: 4, max: 24 }),
        (_yOffset: number) => {
          // Find the GlassCard that has "!p-0" (the one wrapping the inbound list)
          // It should use overflow-x-hidden, not overflow-hidden
          const glassCardPattern = /className="!p-0\s+overflow-x-hidden"/;
          const hasCorrectOverflow = glassCardPattern.test(inboundsSource);
          expect(hasCorrectOverflow).toBe(true);
        },
      ),
      { numRuns: 10 },
    );
  });
});
