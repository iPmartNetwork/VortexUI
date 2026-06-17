import { cn } from "@/lib/utils";

interface GeoPoint {
  country: string;
  connections: number;
  bytes_down: number;
}

interface WorldMapProps {
  data: GeoPoint[];
  className?: string;
}

/**
 * Simple SVG world map heatmap. Uses a simplified world outline and
 * overlays colored circles at approximate country center coordinates.
 * For a production app, use react-simple-maps for full interaction.
 */
export function WorldMap({ data, className }: WorldMapProps) {
  const maxConn = Math.max(...data.map(d => d.connections), 1);

  return (
    <div className={cn("relative rounded-xl bg-surface-2/30 p-4 overflow-hidden", className)}>
      {/* Simplified world SVG background */}
      <svg viewBox="0 0 1000 500" className="w-full h-auto opacity-20" preserveAspectRatio="xMidYMid">
        {/* Rough continent outlines */}
        <ellipse cx="250" cy="200" rx="120" ry="100" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="520" cy="180" rx="130" ry="120" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="750" cy="230" rx="100" ry="90" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="200" cy="350" rx="60" ry="80" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="530" cy="380" rx="50" ry="50" fill="hsl(var(--fg-subtle))" opacity="0.3" />
        <ellipse cx="830" cy="380" rx="70" ry="50" fill="hsl(var(--fg-subtle))" opacity="0.3" />
      </svg>

      {/* Data points overlay */}
      <svg viewBox="0 0 1000 500" className="absolute inset-0 w-full h-full p-4" preserveAspectRatio="xMidYMid">
        {data.slice(0, 20).map((point, i) => {
          const coords = getCountryCoords(point.country);
          if (!coords) return null;
          const size = Math.max(8, (point.connections / maxConn) * 40);
          const opacity = 0.3 + (point.connections / maxConn) * 0.7;
          return (
            <g key={i}>
              <circle cx={coords.x} cy={coords.y} r={size} fill="hsl(var(--primary))" opacity={opacity * 0.3} />
              <circle cx={coords.x} cy={coords.y} r={size * 0.5} fill="hsl(var(--primary))" opacity={opacity} className="animate-pulse" />
            </g>
          );
        })}
      </svg>

      {/* Legend */}
      <div className="absolute bottom-3 start-3 flex items-center gap-3 text-[10px] text-fg-subtle">
        <div className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-primary/30" /> Low</div>
        <div className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-primary/60" /> Medium</div>
        <div className="flex items-center gap-1"><span className="h-2 w-2 rounded-full bg-primary" /> High</div>
      </div>
    </div>
  );
}

// Approximate country center coordinates on our 1000x500 map
const COUNTRY_COORDS: Record<string, { x: number; y: number }> = {
  IR: { x: 590, y: 205 }, US: { x: 200, y: 180 }, DE: { x: 485, y: 155 },
  NL: { x: 478, y: 148 }, FR: { x: 468, y: 162 }, GB: { x: 460, y: 142 },
  TR: { x: 545, y: 175 }, RU: { x: 620, y: 130 }, CN: { x: 730, y: 195 },
  JP: { x: 810, y: 185 }, KR: { x: 785, y: 185 }, IN: { x: 670, y: 235 },
  AE: { x: 600, y: 225 }, SG: { x: 720, y: 285 }, AU: { x: 820, y: 380 },
  BR: { x: 280, y: 340 }, CA: { x: 190, y: 140 }, IT: { x: 490, y: 170 },
  ES: { x: 455, y: 175 }, SE: { x: 495, y: 120 }, FI: { x: 520, y: 115 },
  PL: { x: 505, y: 150 }, UA: { x: 540, y: 150 }, EG: { x: 535, y: 220 },
  SA: { x: 580, y: 230 }, PK: { x: 645, y: 215 }, ID: { x: 740, y: 295 },
  VN: { x: 730, y: 250 }, TH: { x: 715, y: 255 }, MY: { x: 720, y: 280 },
  ZA: { x: 530, y: 395 }, NG: { x: 480, y: 270 }, KE: { x: 550, y: 290 },
  AR: { x: 260, y: 400 }, MX: { x: 170, y: 235 }, CL: { x: 245, y: 395 },
};

function getCountryCoords(code: string): { x: number; y: number } | null {
  return COUNTRY_COORDS[code?.toUpperCase()] ?? null;
}
