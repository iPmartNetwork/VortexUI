import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';
import translations, { langMeta, type Lang } from '../i18n/translations';

interface LangContextType {
  lang: Lang;
  setLang: (l: Lang) => void;
  t: (key: string) => string;
  dir: 'ltr' | 'rtl';
  isRTL: boolean;
}

const LangContext = createContext<LangContextType>({
  lang: 'en',
  setLang: () => {},
  t: (k) => k,
  dir: 'ltr',
  isRTL: false,
});

export function LangProvider({ children }: { children: ReactNode }) {
  const [lang, setLangState] = useState<Lang>(() => {
    if (typeof window !== 'undefined') {
      return (localStorage.getItem('vortex-lang') as Lang) || 'en';
    }
    return 'en';
  });

  const setLang = (l: Lang) => {
    setLangState(l);
    localStorage.setItem('vortex-lang', l);
  };

  const dir = langMeta[lang].dir;
  const isRTL = dir === 'rtl';

  useEffect(() => {
    document.documentElement.setAttribute('dir', dir);
    document.documentElement.setAttribute('lang', lang);
  }, [lang, dir]);

  const t = (key: string): string => {
    return translations[lang]?.[key] || translations.en[key] || key;
  };

  return (
    <LangContext.Provider value={{ lang, setLang, t, dir, isRTL }}>
      {children}
    </LangContext.Provider>
  );
}

export const useLang = () => useContext(LangContext);
