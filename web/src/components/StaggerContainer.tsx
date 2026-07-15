import { motion } from "framer-motion";
import { cn } from "@/lib/utils";

interface StaggerContainerProps {
  children: React.ReactNode;
  className?: string;
  /** Delay between each child (seconds). Default 0.05 */
  staggerDelay?: number;
  /** Initial y offset. Default 12 */
  yOffset?: number;
  /** Animation duration. Default 0.35 */
  duration?: number;
}

/**
 * Wraps children (direct children of this component) with staggered
 * fade+slide-up animation. Each child animates in sequence with a delay.
 *
 * @example
 * <StaggerContainer>
 *   <div>First</div>   ← animates at t=0
 *   <div>Second</div>  ← animates at t=0.05
 *   <div>Third</div>   ← animates at t=0.1
 * </StaggerContainer>
 */
export function StaggerContainer({
  children,
  className,
  staggerDelay = 0.05,
  yOffset = 12,
  duration = 0.35,
}: StaggerContainerProps) {
  return (
    <div className={cn(className)}>
      {Array.isArray(children)
        ? (children as React.ReactNode[]).map((child, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: yOffset }}
              animate={{ opacity: 1, y: 0 }}
              transition={{
                delay: i * staggerDelay,
                duration,
                ease: [0.16, 1, 0.3, 1],
              }}
            >
              {child}
            </motion.div>
          ))
        : children}
    </div>
  );
}
