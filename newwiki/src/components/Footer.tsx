import { Heart, ArrowUpRight } from 'lucide-react';
import { GithubIcon, TelegramIcon } from './icons';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

export default function Footer() {
  const { isDark } = useTheme();
  const { t } = useLang();

  const linkGroups = [
    {
      title: t('footer.product'),
      items: [
        { name: t('nav.features'), href: '#features' },
        { name: t('nav.protocols'), href: '#protocols' },
        { name: t('nav.security'), href: '#security' },
        { name: t('nav.comparison'), href: '#comparison' },
        { name: t('nav.install'), href: '#install' },
      ],
    },
    {
      title: t('footer.resources'),
      items: [
        { name: t('nav.documentation'), href: 'https://ipmartnetwork.github.io/VortexUI/', external: true },
        { name: 'API Reference', href: 'https://ipmartnetwork.github.io/VortexUI/12-api-reference/', external: true },
        { name: 'OpenAPI Spec', href: 'https://github.com/iPmartNetwork/VortexUI/blob/master/docs/openapi.yaml', external: true },
        { name: 'Changelog', href: 'https://github.com/iPmartNetwork/VortexUI/blob/master/CHANGELOG.md', external: true },
      ],
    },
    {
      title: t('footer.community'),
      items: [
        { name: t('nav.github'), href: 'https://github.com/iPmartNetwork/VortexUI', external: true },
        { name: t('nav.telegram'), href: 'https://t.me/vortex_ui', external: true },
        { name: 'Discussions', href: 'https://github.com/iPmartNetwork/VortexUI/discussions', external: true },
        { name: 'Bug Reports', href: 'https://github.com/iPmartNetwork/VortexUI/issues', external: true },
      ],
    },
  ];

  return (
    <footer className={`relative border-t ${isDark ? 'border-white/5' : 'border-gray-200'}`}>
      <div className={`absolute inset-0 ${isDark ? 'bg-gradient-to-t from-vortex-900/10 to-transparent' : 'bg-gradient-to-t from-vortex-50/30 to-transparent'}`} />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-12">
          {/* Brand */}
          <div>
            <a href="#" className="flex items-center gap-2 mb-4">
              <span className={`text-xl font-bold ${isDark ? 'text-white' : 'text-gray-900'}`}>Vortex</span>
              <span className="text-xl font-bold text-vortex-500">UI</span>
            </a>
            <p className="text-sm text-themed-muted leading-relaxed mb-6">{t('footer.desc')}</p>
            <div className="flex items-center gap-3">
              <a href="https://github.com/iPmartNetwork/VortexUI" target="_blank" rel="noopener noreferrer"
                className={`w-10 h-10 rounded-xl glass flex items-center justify-center text-themed-muted hover:text-themed transition-all`}>
                <GithubIcon className="w-5 h-5" />
              </a>
              <a href="https://t.me/vortex_ui" target="_blank" rel="noopener noreferrer"
                className={`w-10 h-10 rounded-xl glass flex items-center justify-center text-themed-muted hover:text-themed transition-all`}>
                <TelegramIcon className="w-5 h-5" />
              </a>
            </div>
          </div>

          {linkGroups.map((group) => (
            <div key={group.title}>
              <h4 className={`text-sm font-semibold uppercase tracking-wider mb-4 ${isDark ? 'text-white/80' : 'text-gray-700'}`}>{group.title}</h4>
              <ul className="space-y-3">
                {group.items.map((item) => (
                  <li key={item.name}>
                    <a href={item.href}
                      target={'external' in item && item.external ? '_blank' : undefined}
                      rel={'external' in item && item.external ? 'noopener noreferrer' : undefined}
                      className="text-sm text-themed-muted hover:text-themed transition-colors flex items-center gap-1 group">
                      {item.name}
                      {'external' in item && item.external && <ArrowUpRight className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className={`mt-16 pt-8 border-t flex flex-col sm:flex-row items-center justify-between gap-4 ${isDark ? 'border-white/5' : 'border-gray-200'}`}>
          <div className="text-sm text-themed-faint flex items-center gap-1">
            {t('footer.madeWith')} <Heart className="w-3.5 h-3.5 text-red-500" /> {t('footer.by')}
          </div>
          <div className="text-sm text-themed-faint">
            VortexUI v1.3.1 — {t('footer.openSource')}
          </div>
        </div>
      </div>
    </footer>
  );
}
