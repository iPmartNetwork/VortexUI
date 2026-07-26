import { useState, useRef, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Check, Globe } from "lucide-react";
import { cn } from "@/lib/utils";
import { useI18n } from "@/i18n/i18n";
import { LANG_OPTIONS } from "@/i18n/lang-options";
import { FlagIcon } from "@/components/Flags";
import type { Lang } from "@/i18n/dict";

// ─── Language name in its own script ─────────────────────────

const NATIVE_NAMES: Record<Lang, string> = {
  en: "English",
  fa: "فارسی",
  tr: "Türkçe",
  ar: "العربية",
  ru: "Русский",
  zh: "中文",
  ja: "日本語",
  es: "Español",
};

// ─── Language Switcher ───────────────────────────────────────

interface LanguageSwitcherProps {
  collapsed?: boolean;
}

export function LanguageSwitcher({ collapsed = false }: LanguageSwitcherProps) {
  const { lang, setLang, dir } = useI18n();
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Close on click outside
  useEffect(() => {
    if (!open) return;
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, [open]);

  const currentLang = LANG_OPTIONS.find((l) => l.code === lang);

  function handleSelect(code: Lang) {
    setLang(code);
    setOpen(false);
  }

  if (collapsed) {
    return (
      <div ref={ref} className="relative">
        <button
          type="button"
          onClick={() => setOpen(!open)}
          className={cn(
            "h-8 w-8 mx-auto rounded-lg flex items-center justify-center",
            "text-lg leading-none transition-all duration-150",
            "hover:bg-surface-2/60 active:scale-95",
            open && "bg-surface-2/60",
          )}
          title={currentLang?.label ?? "Language"}
        >
          <div className="h-5 w-[15px] flex-shrink-0">
            <FlagIcon lang={lang} className="h-full w-full" />
          </div>
        </button>

        <AnimatePresence>
          {open && (
            <>
              {/* Arrow */}
              <motion.div
                initial={{ opacity: 0, x: dir === "rtl" ? 4 : -4 }}
                animate={{ opacity: 1, x: 0 }}
                exit={{ opacity: 0, x: dir === "rtl" ? 4 : -4 }}
                transition={{ duration: 0.1 }}
                className={cn(
                  "absolute top-1/2 -translate-y-1/2 z-[60]",
                  dir === "rtl" ? "end-full -me-1.5" : "start-full ms-1.5",
                )}
              >
                <div className="flex flex-col gap-0.5 rounded-xl bg-fg p-1 shadow-xl min-w-[120px]">
                  {LANG_OPTIONS.map((l) => {
                    const active = l.code === lang;
                    return (
                      <button
                        key={l.code}
                        type="button"
                        onClick={() => handleSelect(l.code)}
                        className={cn(
                          "flex items-center gap-2 rounded-lg px-2.5 py-1.5 text-[11px] font-medium transition-colors text-left",
                          active
                            ? "bg-primary/20 text-primary"
                            : "text-bg hover:bg-surface-2/20",
                        )}
                      >
                        <div className="h-4 w-3 flex-shrink-0 rounded-[2px] overflow-hidden shadow-sm">
                          <FlagIcon lang={l.code} className="h-full w-full" />
                        </div>
                        <span className="flex-1">{l.label}</span>
                        {active && <Check size={12} className="flex-shrink-0" />}
                      </button>
                    );
                  })}
                </div>
              </motion.div>
            </>
          )}
        </AnimatePresence>
      </div>
    );
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          "w-full h-9 rounded-xl text-[11px] font-medium flex items-center gap-2.5 px-3",
          "transition-all duration-150 border",
          open
            ? "border-primary/30 bg-primary/5 text-fg"
            : "border-border/60 text-fg-muted hover:text-fg hover:bg-surface-2/40",
        )}
      >
        <Globe size={14} className="flex-shrink-0 opacity-60" />
        <div className="h-5 w-[15px] flex-shrink-0 rounded-[2px] overflow-hidden shadow-sm">
          <FlagIcon lang={lang} className="h-full w-full" />
        </div>
        <span className="flex-1 text-start truncate">{currentLang?.label ?? "Language"}</span>
        <motion.span
          animate={{ rotate: open ? 180 : 0 }}
          transition={{ duration: 0.2 }}
          className="flex-shrink-0 opacity-40"
        >
          <svg width="8" height="5" viewBox="0 0 8 5" fill="none">
            <path d="M1 1L4 4L7 1" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
          </svg>
        </motion.span>
      </button>

      <AnimatePresence>
        {open && (
          <>
            {/* Backdrop */}
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40"
              onClick={() => setOpen(false)}
            />

            {/* Dropdown */}
            <motion.div
              initial={{ opacity: 0, y: -4, scale: 0.96 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -4, scale: 0.96 }}
              transition={{ duration: 0.15, ease: [0.16, 1, 0.3, 1] }}
              className={cn(
                "absolute z-50 w-full min-w-[160px] rounded-xl border bg-bg-elevated border-border/70 shadow-lg overflow-hidden",
                dir === "rtl" ? "end-0" : "start-0",
                "bottom-full mb-1.5", // opens upward
              )}
            >
              <div className="p-1 space-y-0.5">
                {LANG_OPTIONS.map((l) => {
                  const active = l.code === lang;
                  return (
                    <button
                      key={l.code}
                      type="button"
                      onClick={() => handleSelect(l.code)}
                      className={cn(
                        "flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left transition-all duration-100",
                        active
                          ? "bg-primary/10 text-primary font-semibold"
                          : "text-fg-muted hover:text-fg hover:bg-surface-2/50",
                      )}
                    >
                      <div className="h-5 w-[15px] flex-shrink-0 rounded-[2px] overflow-hidden shadow-sm">
                        <FlagIcon lang={l.code} className="h-full w-full" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <span className="text-[12px] block truncate">{l.label}</span>
                        <span className="text-[9px] text-fg-subtle/70 block truncate">
                          {NATIVE_NAMES[l.code]}
                        </span>
                      </div>
                      {active && (
                        <motion.span
                          initial={{ scale: 0 }}
                          animate={{ scale: 1 }}
                          transition={{ type: "spring", stiffness: 500, damping: 25 }}
                        >
                          <Check size={13} className="flex-shrink-0" />
                        </motion.span>
                      )}
                    </button>
                  );
                })}
              </div>

              {/* Decorative gradient line */}
              <div className="h-px bg-gradient-to-r from-transparent via-primary/20 to-transparent mx-3" />
              <div className="px-3 py-1.5 flex items-center justify-center gap-1.5">
                <Globe size={10} className="text-fg-subtle/40" />
                <span className="text-[8px] text-fg-subtle/40 font-medium tracking-wider uppercase">
                  {LANG_OPTIONS.length} languages
                </span>
              </div>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </div>
  );
}
