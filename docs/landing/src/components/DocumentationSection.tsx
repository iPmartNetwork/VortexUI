import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import {
  BookOpen, Settings, Users, Server, Shield, CreditCard,
  Bell, Database, Globe, Terminal, HelpCircle, FileCode,
  Layers, Activity, ArrowRight, Cpu
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

const docSections = [
  { icon: <BookOpen className="w-5 h-5" />, title: 'Introduction', color: 'text-blue-500', href: 'https://ipmartnetwork.github.io/VortexUI/01-introduction/' },
  { icon: <Terminal className="w-5 h-5" />, title: 'Installation', color: 'text-green-500', href: 'https://ipmartnetwork.github.io/VortexUI/02-installation/' },
  { icon: <Cpu className="w-5 h-5" />, title: 'First Steps', color: 'text-cyan-500', href: 'https://ipmartnetwork.github.io/VortexUI/03-first-steps/' },
  { icon: <Activity className="w-5 h-5" />, title: 'Dashboard', color: 'text-vortex-500', href: 'https://ipmartnetwork.github.io/VortexUI/04-dashboard/' },
  { icon: <Users className="w-5 h-5" />, title: 'Users', color: 'text-amber-500', href: 'https://ipmartnetwork.github.io/VortexUI/05-user-management/' },
  { icon: <Server className="w-5 h-5" />, title: 'Nodes', color: 'text-red-500', href: 'https://ipmartnetwork.github.io/VortexUI/06-node-management/' },
  { icon: <Globe className="w-5 h-5" />, title: 'Network', color: 'text-teal-500', href: 'https://ipmartnetwork.github.io/VortexUI/07-network-policy/' },
  { icon: <Shield className="w-5 h-5" />, title: 'Security', color: 'text-rose-500', href: 'https://ipmartnetwork.github.io/VortexUI/08-security-administration/' },
  { icon: <CreditCard className="w-5 h-5" />, title: 'Plans & Payments', color: 'text-emerald-500', href: 'https://ipmartnetwork.github.io/VortexUI/09-plans-payments/' },
  { icon: <Bell className="w-5 h-5" />, title: 'Notifications', color: 'text-yellow-500', href: 'https://ipmartnetwork.github.io/VortexUI/10-notifications/' },
  { icon: <Settings className="w-5 h-5" />, title: 'Settings', color: 'text-purple-500', href: 'https://ipmartnetwork.github.io/VortexUI/11-settings-backup/' },
  { icon: <FileCode className="w-5 h-5" />, title: 'API Reference', color: 'text-indigo-500', href: 'https://ipmartnetwork.github.io/VortexUI/12-api-reference/' },
  { icon: <Layers className="w-5 h-5" />, title: 'Protocols', color: 'text-sky-500', href: 'https://ipmartnetwork.github.io/VortexUI/13-protocols-config/' },
  { icon: <Database className="w-5 h-5" />, title: 'Operations', color: 'text-orange-500', href: 'https://ipmartnetwork.github.io/VortexUI/14-operations-maintenance/' },
  { icon: <HelpCircle className="w-5 h-5" />, title: 'Troubleshooting', color: 'text-pink-500', href: 'https://ipmartnetwork.github.io/VortexUI/15-troubleshooting-faq/' },
];

export default function DocumentationSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  return (
    <section id="docs" className="relative py-32 overflow-hidden">
      <div className={`absolute top-0 left-1/2 -translate-x-1/2 w-[1200px] h-[600px] ${isDark ? 'bg-vortex-500/3' : 'bg-vortex-500/2'} rounded-full blur-3xl`} />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm mb-6 ${isDark ? 'text-vortex-400' : 'text-vortex-600'}`}>
            <BookOpen className="w-4 h-4" /> {t('docs.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('docs.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-vortex-400 to-cyber-400 bg-clip-text text-transparent">{t('docs.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('docs.subtitle')}</p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
          {docSections.map((section, i) => (
            <motion.a key={section.title} href={section.href} target="_blank" rel="noopener noreferrer"
              initial={{ opacity: 0, y: 20 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.3, delay: i * 0.04 }}
              className="glass rounded-xl p-4 card-hover group block">
              <div className={`${section.color} mb-3`}>{section.icon}</div>
              <div className={`text-sm font-semibold mb-1 flex items-center gap-1 ${isDark ? 'text-white/90' : 'text-gray-800'}`}>
                {section.title}
                <ArrowRight className="w-3 h-3 opacity-0 group-hover:opacity-100 group-hover:translate-x-1 rtl:group-hover:-translate-x-1 transition-all text-themed-muted" />
              </div>
            </motion.a>
          ))}
        </div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.5, delay: 0.6 }} className="mt-12 text-center">
          <a href="https://ipmartnetwork.github.io/VortexUI/" target="_blank" rel="noopener noreferrer"
            className={`inline-flex items-center gap-2 px-8 py-4 rounded-2xl text-base font-semibold glass border transition-all group ${
              isDark ? 'text-white border-vortex-500/20 hover:border-vortex-500/40 hover:bg-vortex-500/5' : 'text-gray-800 border-vortex-200 hover:border-vortex-400 hover:bg-vortex-50'
            }`}>
            <BookOpen className="w-5 h-5 text-vortex-500" />
            {t('docs.openFull')}
            <ArrowRight className="w-4 h-4 text-themed-muted group-hover:translate-x-1 rtl:group-hover:-translate-x-1 transition-transform" />
          </a>
        </motion.div>
      </div>
    </section>
  );
}
