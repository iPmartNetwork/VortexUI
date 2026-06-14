import { useRef } from "react";

// highlight wraps JSON tokens in colored spans. Keys keep the default fg color,
// string values are amber, numbers green, booleans/null blue — 3x-ui style.
function highlight(src: string): string {
  const esc = src.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  return esc.replace(
    /("(?:\\u[\da-fA-F]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(?:true|false|null)\b|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?)/g,
    (m) => {
      let cls = "text-emerald-500"; // number
      if (m.startsWith('"')) cls = /:\s*$/.test(m) ? "text-fg" : "text-amber-500";
      else if (/^(true|false|null)$/.test(m)) cls = "text-sky-400";
      return `<span class="${cls}">${m}</span>`;
    },
  );
}

// JsonCodeEditor is a dependency-free code editor with a line-number gutter and
// live syntax highlighting (a highlighted <pre> behind a transparent textarea).
export function JsonCodeEditor({
  value,
  onChange,
  rows = 14,
}: {
  value: string;
  onChange: (v: string) => void;
  rows?: number;
}) {
  const taRef = useRef<HTMLTextAreaElement>(null);
  const preRef = useRef<HTMLPreElement>(null);
  const gutterRef = useRef<HTMLDivElement>(null);
  const lines = value.split("\n");

  function syncScroll() {
    if (preRef.current && taRef.current) {
      preRef.current.scrollTop = taRef.current.scrollTop;
      preRef.current.scrollLeft = taRef.current.scrollLeft;
    }
    if (gutterRef.current && taRef.current) gutterRef.current.scrollTop = taRef.current.scrollTop;
  }

  function onKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (e.key === "Tab") {
      e.preventDefault();
      const ta = e.currentTarget;
      const { selectionStart: s, selectionEnd: end } = ta;
      const next = value.slice(0, s) + "  " + value.slice(end);
      onChange(next);
      requestAnimationFrame(() => ta.setSelectionRange(s + 2, s + 2));
    }
  }

  const h = rows * 1.25; // rem; leading-5 == 1.25rem per line

  return (
    <div className="relative flex overflow-hidden rounded-lg border border-border/60 bg-surface-2/40 font-mono text-xs" dir="ltr">
      <div
        ref={gutterRef}
        className="select-none overflow-hidden border-e border-border/50 bg-surface-2/30 py-3 text-end text-fg-subtle/60"
        style={{ height: `${h}rem` }}
        aria-hidden
      >
        {lines.map((_, i) => (
          <div key={i} className="px-2 leading-5">{i + 1}</div>
        ))}
      </div>
      <div className="relative flex-1">
        <pre
          ref={preRef}
          aria-hidden
          className="pointer-events-none m-0 overflow-auto whitespace-pre px-3 py-3 leading-5"
          style={{ height: `${h}rem` }}
        >
          <code dangerouslySetInnerHTML={{ __html: highlight(value) + "\n" }} />
        </pre>
        <textarea
          ref={taRef}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          onScroll={syncScroll}
          onKeyDown={onKeyDown}
          spellCheck={false}
          autoCapitalize="off"
          autoCorrect="off"
          className="absolute inset-0 w-full resize-none overflow-auto whitespace-pre bg-transparent px-3 py-3 leading-5 outline-none"
          style={{ height: `${h}rem`, color: "transparent", caretColor: "#888" }}
        />
      </div>
    </div>
  );
}
