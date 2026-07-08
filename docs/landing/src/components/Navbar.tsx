import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Menu, X, BookOpen, MessageCircle, Sun, Moon, ChevronDown } from 'lucide-react';
import { GithubIcon } from './icons';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';
import { langMeta, type Lang } from '../i18n/translations';

export default function Navbar() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);
  const [langOpen, setLangOpen] = useState(false);
  const { isDark, toggle: toggleTheme } = useTheme();
  const { t, lang, setLang, isRTL } = useLang();

  const navLinks = [
    { name: t('nav.features'), href: '#features' },
    { name: t('nav.architecture'), href: '#architecture' },
    { name: t('nav.protocols'), href: '#protocols' },
    { name: t('nav.security'), href: '#security' },
    { name: t('nav.comparison'), href: '#comparison' },
    { name: t('nav.install'), href: '#install' },
  ];

  useEffect(() => {
    const handleScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  useEffect(() => {
    const close = () => setLangOpen(false);
    if (langOpen) {
      document.addEventListener('click', close);
      return () => document.removeEventListener('click', close);
    }
  }, [langOpen]);

  return (
    <>
      <motion.nav
        initial={{ y: -100 }}
        animate={{ y: 0 }}
        transition={{ duration: 0.6, ease: 'easeOut' }}
        className={`fixed top-0 left-0 right-0 z-50 transition-all duration-500 ${
          scrolled
            ? isDark
              ? 'glass-strong shadow-lg shadow-vortex-900/20'
              : 'bg-white/80 backdrop-blur-xl shadow-lg shadow-black/5 border-b border-vortex-100'
            : 'bg-transparent'
        }`}
      >
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16 lg:h-20">
            {/* Logo - text only */}
            <a href="#" className="flex items-center gap-2 group">
              <div>
                <span className={`text-xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>Vortex</span>
                <span className="text-xl font-bold bg-gradient-to-r from-vortex-500 to-cyber-400 bg-clip-text text-transparent">UI</span>
                <span className={`hidden sm:inline ms-2 text-xs font-mono ${isDark ? 'text-white/30' : 'text-gray-400'}`}>v1.3.1</span>
              </div>
            </a>

            {/* Desktop Nav */}
            <div className="hidden lg:flex items-center gap-1">
              {navLinks.map((link) => (
                <a
                  key={link.href}
                  href={link.href}
                  className={`px-4 py-2 text-sm rounded-lg transition-all duration-300 ${
                    isDark
                      ? 'text-white/60 hover:text-white hover:bg-white/5'
                      : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  {link.name}
                </a>
              ))}
            </div>

            {/* Right Actions */}
            <div className="hidden lg:flex items-center gap-2">
              {/* Language Selector */}
              <div className="relative">
                <button
                  onClick={(e) => { e.stopPropagation(); setLangOpen(!langOpen); }}
                  className={`flex items-center gap-1.5 px-3 py-2 rounded-xl text-sm transition-all ${
                    isDark ? 'text-white/60 hover:text-white hover:bg-white/5' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100'
                  }`}
                >
                  <span>{langMeta[lang].flag}</span>
                  <span className="hidden xl:inline">{langMeta[lang].label}</span>
                  <ChevronDown className="w-3 h-3" />
                </button>
                <AnimatePresence>
                  {langOpen && (
                    <motion.div
                      initial={{ opacity: 0, y: -8, scale: 0.95 }}
                      animate={{ opacity: 1, y: 0, scale: 1 }}
                      exit={{ opacity: 0, y: -8, scale: 0.95 }}
                      className={`absolute top-full mt-2 ${isRTL ? 'right-0' : 'left-0'} min-w-[160px] rounded-xl overflow-hidden z-50 ${
                        isDark ? 'glass-strong' : 'bg-white shadow-xl border border-gray-200'
                      }`}
                    >
                      {(Object.keys(langMeta) as Lang[]).map((l) => (
                        <button
                          key={l}
                          onClick={(e) => { e.stopPropagation(); setLang(l); setLangOpen(false); }}
                          className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-colors ${
                            lang === l
                              ? isDark ? 'bg-vortex-500/20 text-vortex-400' : 'bg-vortex-50 text-vortex-600'
                              : isDark ? 'text-white/60 hover:bg-white/5 hover:text-white' : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                          }`}
                        >
                          <span>{langMeta[l].flag}</span>
                          <span>{langMeta[l].label}</span>
                        </button>
                      ))}
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>

              {/* Theme Toggle */}
              <button
                onClick={toggleTheme}
                className={`p-2.5 rounded-xl transition-all ${
                  isDark
                    ? 'text-white/60 hover:text-amber-400 hover:bg-white/5'
                    : 'text-gray-500 hover:text-vortex-600 hover:bg-gray-100'
                }`}
                title={isDark ? 'Light mode' : 'Dark mode'}
              >
                {isDark ? <Sun className="w-4.5 h-4.5" /> : <Moon className="w-4.5 h-4.5" />}
              </button>

              <a
                href="01-introduction/"
                rel="noopener noreferrer"
                className={`flex items-center gap-2 px-3 py-2 text-sm transition-colors ${
                  isDark ? 'text-white/70 hover:text-white' : 'text-gray-500 hover:text-gray-900'
                }`}
              >
                <BookOpen className="w-4 h-4" />
                {t('nav.docs')}
              </a>
              <a
                href="https://github.com/iPmartNetwork/VortexUI"
                target="_blank"
                rel="noopener noreferrer"
                className={`flex items-center gap-2 px-3 py-2 text-sm transition-colors ${
                  isDark ? 'text-white/70 hover:text-white' : 'text-gray-500 hover:text-gray-900'
                }`}
              >
                <GithubIcon className="w-4 h-4" />
                {t('nav.github')}
              </a>
              <a
                href="#install"
                className="relative group px-5 py-2.5 rounded-xl text-sm font-semibold text-white overflow-hidden"
              >
                <div className="absolute inset-0 bg-gradient-to-r from-vortex-600 to-vortex-500 group-hover:from-vortex-500 group-hover:to-cyber-500 transition-all duration-500" />
                <span className="relative z-10">{t('nav.getStarted')}</span>
              </a>
            </div>

            {/* Mobile: theme + lang + menu */}
            <div className="flex lg:hidden items-center gap-2">
              <button
                onClick={toggleTheme}
                className={`p-2 rounded-lg ${isDark ? 'text-white/60' : 'text-gray-500'}`}
              >
                {isDark ? <Sun className="w-5 h-5" /> : <Moon className="w-5 h-5" />}
              </button>
              <button
                onClick={() => setMobileOpen(!mobileOpen)}
                className={`p-2 ${isDark ? 'text-white/70' : 'text-gray-600'}`}
              >
                {mobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
              </button>
            </div>
          </div>
        </div>
      </motion.nav>

      {/* Mobile Menu */}
      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            className={`fixed inset-0 z-40 pt-20 px-6 lg:hidden ${
              isDark ? 'bg-[#030014]/95 backdrop-blur-xl' : 'bg-white/95 backdrop-blur-xl'
            }`}
          >
            <div className="flex flex-col gap-2">
              {navLinks.map((link, i) => (
                <motion.a
                  key={link.href}
                  href={link.href}
                  initial={{ opacity: 0, x: isRTL ? 20 : -20 }}
                  animate={{ opacity: 1, x: 0 }}
                  transition={{ delay: i * 0.05 }}
                  onClick={() => setMobileOpen(false)}
                  className={`px-4 py-3 text-lg rounded-xl transition-all ${
                    isDark ? 'text-white/80 hover:bg-white/5' : 'text-gray-700 hover:bg-gray-100'
                  }`}
                >
                  {link.name}
                </motion.a>
              ))}

              {/* Language selector in mobile */}
              <div className={`border-t mt-4 pt-4 ${isDark ? 'border-white/10' : 'border-gray-200'}`}>
                <div className="flex flex-wrap gap-2 mb-4">
                  {(Object.keys(langMeta) as Lang[]).map((l) => (
                    <button
                      key={l}
                      onClick={() => { setLang(l); }}
                      className={`flex items-center gap-2 px-3 py-2 rounded-lg text-sm ${
                        lang === l
                          ? 'bg-vortex-500/20 text-vortex-400 border border-vortex-500/30'
                          : isDark ? 'text-white/50 hover:text-white glass' : 'text-gray-500 hover:text-gray-900 bg-gray-100'
                      }`}
                    >
                      <span>{langMeta[l].flag}</span>
                      {langMeta[l].label}
                    </button>
                  ))}
                </div>
                <div className="flex flex-col gap-2">
                  <a href="01-introduction/" rel="noopener noreferrer"
                    className={`flex items-center gap-3 px-4 py-3 ${isDark ? 'text-white/70' : 'text-gray-600'}`}
                    onClick={() => setMobileOpen(false)}>
                    <BookOpen className="w-5 h-5" /> {t('nav.documentation')}
                  </a>
                  <a href="https://github.com/iPmartNetwork/VortexUI" target="_blank" rel="noopener noreferrer"
                    className={`flex items-center gap-3 px-4 py-3 ${isDark ? 'text-white/70' : 'text-gray-600'}`}
                    onClick={() => setMobileOpen(false)}>
                    <GithubIcon className="w-5 h-5" /> {t('nav.github')}
                  </a>
                  <a href="https://t.me/vortex_ui" target="_blank" rel="noopener noreferrer"
                    className={`flex items-center gap-3 px-4 py-3 ${isDark ? 'text-white/70' : 'text-gray-600'}`}
                    onClick={() => setMobileOpen(false)}>
                    <MessageCircle className="w-5 h-5" /> {t('nav.telegram')}
                  </a>
                </div>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
