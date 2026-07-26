import type { Lang } from "@/i18n/dict";

// ════════════════════════════════════════════════════════════════
// SVG Flag Icons — 4:3 aspect ratio, clean vector graphics
// Each flag is a self-contained React component with viewBox="0 0 20 15"
// ════════════════════════════════════════════════════════════════

function FlagBase({ children, className }: { children: React.ReactNode; className?: string }) {
  return (
    <svg
      viewBox="0 0 20 15"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      {children}
    </svg>
  );
}

// ── United Kingdom ─────────────────────────────────────────────
function FlagGB({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#012169" />
      <path d="M0 0L20 15M20 0L0 15" stroke="#fff" strokeWidth="3" />
      <path d="M0 0L20 15M20 0L0 15" stroke="#C8102E" strokeWidth="1.5" />
      <rect x="8.5" y="0" width="3" height="15" fill="#fff" />
      <rect x="9.25" y="0" width="1.5" height="15" fill="#C8102E" />
      <rect x="0" y="6" width="20" height="3" fill="#fff" />
      <rect x="0" y="6.75" width="20" height="1.5" fill="#C8102E" />
    </FlagBase>
  );
}

// ── Iran ───────────────────────────────────────────────────────
function FlagIR({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#239F40" />
      <rect y="0" width="20" height="4.5" fill="#DA0000" />
      <rect y="10.5" width="20" height="4.5" fill="#DA0000" />
      {/* Simplified tulip emblem in center */}
      <g transform="translate(10,7.5) scale(0.4)" fill="#DA0000">
        <path d="M0-6C-2-4-4-2-4 1C-4 3-2 5 0 6C2 5 4 3 4 1C4-2 2-4 0-6Z" />
        <path d="M0-3L-2-1L0 1L2-1Z" fill="#239F40" />
      </g>
      {/* Takbir stripes */}
      {[-1, 0, 1].map((i) => (
        <rect key={i} x={9.5 + i * 7} y="0.5" width="1.2" height="2" fill="#fff" rx="0.2" />
      ))}
      {[-1, 0, 1].map((i) => (
        <rect key={i} x={9.5 + i * 7} y="12.5" width="1.2" height="2" fill="#fff" rx="0.2" />
      ))}
    </FlagBase>
  );
}

// ── Turkey ─────────────────────────────────────────────────────
function FlagTR({ className }: { className?: string }) {
  const cx = 16;
  const cy = 7.5;
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#E30A17" />
      {/* Crescent */}
      <circle cx={cx} cy={cy} r="4.5" fill="#fff" />
      <circle cx={cx + 1} cy={cy} r="4" fill="#E30A17" />
      {/* Star */}
      <g transform={`translate(${cx - 1.8}, ${cy - 0.5}) scale(0.35)`} fill="#fff">
        <polygon points="0,-5 1.2,-1.5 5,-1.5 2,0.5 3.2,4 0,1.5 -3.2,4 -2,0.5 -5,-1.5 -1.2,-1.5" />
      </g>
    </FlagBase>
  );
}

// ── Saudi Arabia ───────────────────────────────────────────────
function FlagSA({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#006C35" />
      {/* Simplified Shahada script representation */}
      <text
        x="10"
        y="9.5"
        textAnchor="middle"
        fill="#fff"
        fontSize="7.5"
        fontFamily="serif"
        fontWeight="bold"
      >
        لا اله الا الله
      </text>
      {/* Sword */}
      <line x1="10" y1="10.5" x2="10" y2="13" stroke="#fff" strokeWidth="0.8" />
    </FlagBase>
  );
}

// ── Russia ─────────────────────────────────────────────────────
function FlagRU({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#fff" />
      <rect y="0" width="20" height="5" fill="#0039A6" />
      <rect y="10" width="20" height="5" fill="#D52B1E" />
    </FlagBase>
  );
}

// ── China ──────────────────────────────────────────────────────
function FlagCN({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#DE2910" />
      {/* Large star */}
      <g transform="translate(4.5, 5.5) scale(0.35)" fill="#FFDE00">
        <polygon points="0,-6 1.5,-2 5.5,-2 2.5,0.5 3.5,4.5 0,2 -3.5,4.5 -2.5,0.5 -5.5,-2 -1.5,-2" />
      </g>
      {/* Small stars */}
      {[
        [2.5, 3],
        [3, 4.5],
        [2.5, 6],
        [1.5, 5.5],
      ].map(([x, y], i) => (
        <g key={i} transform={`translate(${x}, ${y}) scale(0.18)`} fill="#FFDE00">
          <polygon points="0,-6 1.5,-2 5.5,-2 2.5,0.5 3.5,4.5 0,2 -3.5,4.5 -2.5,0.5 -5.5,-2 -1.5,-2" />
        </g>
      ))}
    </FlagBase>
  );
}

// ── Japan ──────────────────────────────────────────────────────
function FlagJP({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#fff" />
      <circle cx="10" cy="7.5" r="4.5" fill="#BC002D" />
    </FlagBase>
  );
}

// ── Spain ──────────────────────────────────────────────────────
function FlagES({ className }: { className?: string }) {
  return (
    <FlagBase className={className}>
      <rect width="20" height="15" rx="1.5" fill="#C60B1E" />
      <rect y="3.75" width="20" height="7.5" fill="#FFC400" />
      {/* Simplified coat of arms */}
      <g transform="translate(10, 7.5) scale(0.5)">
        <rect x="-3" y="-2" width="6" height="4" rx="0.5" fill="#C60B1E" />
        <rect x="-2.5" y="-1.5" width="5" height="3" rx="0.3" fill="#FFC400" />
        <rect x="-0.5" y="-1" width="1" height="2" fill="#C60B1E" />
        <rect x="-1.5" y="-0.5" width="3" height="1" fill="#C60B1E" />
      </g>
    </FlagBase>
  );
}

// ════════════════════════════════════════════════════════════════
// Flag Map — access by language code
// ════════════════════════════════════════════════════════════════

export const FLAG_COMPONENTS: Record<Lang, React.ComponentType<{ className?: string }>> = {
  en: FlagGB,
  fa: FlagIR,
  tr: FlagTR,
  ar: FlagSA,
  ru: FlagRU,
  zh: FlagCN,
  ja: FlagJP,
  es: FlagES,
};

// ════════════════════════════════════════════════════════════════
// FlagWrapper — renders the correct flag by language code
// ════════════════════════════════════════════════════════════════

export function FlagIcon({ lang, className }: { lang: Lang; className?: string }) {
  const Flag = FLAG_COMPONENTS[lang];
  if (!Flag) return null;
  return <Flag className={className} />;
}
