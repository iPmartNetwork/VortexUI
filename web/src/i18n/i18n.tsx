import { createContext, useCallback, useContext, useEffect, useMemo, useState } from "react";
import { dict, type Lang, type TKey } from "./dict";

interface I18nState {
  lang: Lang;
  dir: "ltr" | "rtl";
  setLang: (l: Lang) => void;
  t: (key: TKey) => string;
}

const I18nContext = createContext<I18nState | null>(null);
const KEY = "vortex.lang";

export function I18nProvider({ children }: { children: React.ReactNode }) {
  const [lang, setLangState] = useState<Lang>(() => (localStorage.getItem(KEY) as Lang) || "en");
  const dir = lang === "fa" ? "rtl" : "ltr";

  useEffect(() => {
    document.documentElement.lang = lang;
    document.documentElement.dir = dir;
  }, [lang, dir]);

  const setLang = useCallback((l: Lang) => {
    localStorage.setItem(KEY, l);
    setLangState(l);
  }, []);

  const t = useCallback((key: TKey) => dict[lang][key] ?? dict.en[key] ?? key, [lang]);

  const value = useMemo<I18nState>(() => ({ lang, dir, setLang, t }), [lang, dir, setLang, t]);
  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
}

export function useI18n(): I18nState {
  const ctx = useContext(I18nContext);
  if (!ctx) throw new Error("useI18n must be used within I18nProvider");
  return ctx;
}
