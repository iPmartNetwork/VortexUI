import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  Bell,
  Search,
  Moon,
  Sun,
  Command,
  Smartphone,
  Globe,
  Sparkles,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/auth/auth";
import { useTheme } from "@/theme/theme";
import { useI18n } from "@/i18n/i18n";
import { openCommandPalette } from "@/lib/commandPalette";
import { LANG_OPTIONS } from "@/i18n/lang-options";
import type { Lang } from "@/i18n/dict";
import { TelemetryTicker } from "@/components/TelemetryTicker";
import { useVersion } from "@/api/hooks";

export function AppHeader() {
  const { session, sudo } = useAuth();
  const { resolved, toggle: toggleTheme } = useTheme();
  const { t, lang, setLang } = useI18n();
  const navigate = useNavigate();
  const panelVersion = useVersion().data;
  const [searchFocused, setSearchFocused] = useState(false);
  const [langOpen, setLangOpen] = useState(false);

  const username = session?.admin.username ?? "Admin";
  const initials = username.slice(0, 2).toUpperCase();
  const currentLang = LANG_OPTIONS.find((l) => l.code === lang);

  return (
    <header className="sticky top-0 z-30 h-14 border-b border-border/50 bg-bg/85 backdrop-blur-xl flex items-center gap-3 px-4 md:px-5 transition-colors duration-300">
      {/* Search */}
      <div
        className={cn(
          "relative w-full max-w-[200px] md:max-w-xs flex-shrink-0 ms-10 md:ms-0 transition-all duration-200",
          searchFocused && "md:max-w-sm",
        )}
      >
        <Search
          size={14}
          className="absolute start-3 top-1/2 -translate-y-1/2 text-fg-subtle pointer-events-none"
        />
        <input
          type="text"
          readOnly
          onFocus={() => {
            setSearchFocused(true);
            openCommandPalette();
          }}
          onBlur={() => setSearchFocused(false)}
          placeholder={t("shell.searchShortcut")}
          className="w-full h-8 rounded-lg bg-surface-2/60 border border-border/70 ps-8 pe-12 text-xs text-fg placeholder:text-fg-subtle/60 focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/10 transition-all cursor-pointer"
        />
        <kbd className="hidden sm:inline-flex absolute end-2.5 top-1/2 -translate-y-1/2 items-center gap-0.5 rounded-md border border-border/50 bg-surface/50 px-1.5 py-0.5 text-[9px] text-fg-subtle font-mono shadow-sm">
          <Command size={9} />K
        </kbd>
      </div>

      {/* Center telemetry */}
      <TelemetryTicker />

      {/* Right controls */}
      <div className="flex items-center gap-1 flex-shrink-0">
        {/* Language */}
        <div className="relative">
          <button
            type="button"
            onClick={() => setLangOpen(!langOpen)}
            className="flex items-center gap-1 px-2 py-1.5 rounded-lg text-xs font-medium text-fg-muted hover:text-fg hover:bg-surface-2/60 transition-all"
          >
            <Globe size={14} />
            <span className="hidden sm:inline text-[11px] font-semibold">
              {currentLang?.code.toUpperCase() ?? "EN"}
            </span>
          </button>
          {langOpen && (
            <>
              <div
                className="fixed inset-0 z-40"
                onClick={() => setLangOpen(false)}
              />
              <div className="absolute end-0 top-full z-50 mt-2 w-36 rounded-xl border border-border/50 bg-bg-elevated/95 p-1 shadow-xl backdrop-blur-xl animate-scale-in">
                {LANG_OPTIONS.map((l) => (
                  <button
                    key={l.code}
                    type="button"
                    onClick={() => {
                      setLang(l.code as Lang);
                      setLangOpen(false);
                    }}
                    className={cn(
                      "flex w-full items-center gap-2 rounded-lg px-3 py-2 text-xs font-medium transition",
                      lang === l.code
                        ? "bg-primary/10 text-primary"
                        : "text-fg-muted hover:bg-surface-2/60 hover:text-fg",
                    )}
                  >
                    <span className="w-5 text-[10px] font-bold text-fg-subtle">
                      {l.code.toUpperCase()}
                    </span>
                    {l.label}
                  </button>
                ))}
              </div>
            </>
          )}
        </div>

        {/* Theme toggle */}
        <button
          type="button"
          onClick={toggleTheme}
          className="h-8 w-8 rounded-lg flex items-center justify-center text-fg-muted hover:text-fg hover:bg-surface-2/60 transition-all"
          aria-label="Toggle theme"
        >
          {resolved === "dark" ? (
            <Sun size={15} />
          ) : (
            <Moon size={15} />
          )}
        </button>

        {/* Portal link */}
        <button
          type="button"
          onClick={() => navigate("/portal/login")}
          className="hidden md:flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium text-accent hover:bg-accent/10 transition-all border border-transparent hover:border-accent/25"
        >
          <Smartphone size={14} />
          <span>{t("shell.userPortal")}</span>
        </button>

        {/* Version badge */}
        {panelVersion && (
          <span className="hidden md:inline-flex items-center rounded-lg bg-primary/8 text-primary border border-primary/20 px-2 py-1 text-[10px] font-bold tracking-wide shadow-sm">
            <Sparkles size={10} className="me-1" />
            PRO v{panelVersion}
          </span>
        )}

        {/* Notifications */}
        <button
          type="button"
          className="relative h-8 w-8 rounded-lg flex items-center justify-center text-fg-muted hover:text-fg hover:bg-surface-2/60 transition-all"
          aria-label="Notifications"
        >
          <Bell size={15} />
          <span className="absolute -end-0.5 -top-0.5 h-2 w-2 rounded-full bg-danger shadow-[0_0_6px] shadow-danger/50" />
        </button>

        {/* User profile */}
        <div className="hidden sm:flex items-center gap-2 ms-1 ps-2 border-s border-border/50">
          <div className="relative">
            <div className="h-7 w-7 rounded-lg grad-bg flex items-center justify-center text-primary-fg text-[10px] font-bold shadow-glow-sm">
              {initials}
            </div>
            <div className="absolute -bottom-0.5 -end-0.5 h-2 w-2 rounded-full bg-success border-2 border-bg" />
          </div>
          <div className="hidden lg:block">
            <p className="text-xs font-semibold text-fg leading-none">
              {username}
            </p>
            <p className="text-[10px] text-fg-subtle mt-0.5 truncate">
              {sudo ? t("shell.superAdmin") : t("shell.reseller")}
            </p>
          </div>
        </div>
      </div>
    </header>
  );
}
