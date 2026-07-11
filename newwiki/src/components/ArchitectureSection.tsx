import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import { Monitor, Server, Database, HardDrive, Globe, Shield, Cpu } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

const techStack = [
  { label: 'Backend', tech: 'Go 1.26, Echo, gRPC, sqlc, pgx' },
  { label: 'Frontend', tech: 'React 18, TypeScript 5.6, Tailwind CSS' },
  { label: 'Database', tech: 'PostgreSQL 16 + TimescaleDB' },
  { label: 'Cache', tech: 'Redis 7' },
  { label: 'Proxy Cores', tech: 'Xray-core, sing-box' },
  { label: 'Web Server', tech: 'Caddy (auto HTTPS)' },
  { label: 'Transport', tech: 'gRPC + mTLS (panel ↔ nodes)' },
  { label: 'Monitoring', tech: 'Prometheus + Grafana' },
];

export default function ArchitectureSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-100px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  const layers = [
    { title: t('arch.clients'), color: 'from-cyan-500 to-blue-600', icon: <Monitor className="w-5 h-5" />, items: ['Browser / PWA', 'User Portal & Shop', 'Clash / sing-box / v2rayNG'] },
    { title: t('arch.webLayer'), color: 'from-vortex-500 to-purple-600', icon: <Globe className="w-5 h-5" />, items: ['Caddy — HTTPS + SPA + DoH'] },
    { title: t('arch.controlPlane'), color: 'from-amber-500 to-orange-600', icon: <Cpu className="w-5 h-5" />, items: ['Panel API — Go 1.26', 'SSE — Live Events', 'Reality Scanner', 'Auto-Migration', 'Reseller Platform'] },
    { title: t('arch.dataLayer'), color: 'from-green-500 to-emerald-600', icon: <Database className="w-5 h-5" />, items: ['PostgreSQL + TimescaleDB', 'Redis — cache + sessions'] },
    { title: t('arch.nodeFleet'), color: 'from-red-500 to-rose-600', icon: <Server className="w-5 h-5" />, items: ['Local Node', 'Remote Node 1 — mTLS', 'Remote Node N — mTLS'] },
    { title: t('arch.federation'), color: 'from-sky-500 to-indigo-600', icon: <Shield className="w-5 h-5" />, items: ['Peer Panel 1', 'Peer Panel 2'] },
  ];

  return (
    <section id="architecture" className="relative py-32 overflow-hidden">
      <div className={`absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[1000px] h-[600px] ${isDark ? 'bg-vortex-500/3' : 'bg-vortex-500/2'} rounded-full blur-3xl`} />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm mb-6 ${isDark ? 'text-cyber-400' : 'text-cyan-600'}`}>
            <HardDrive className="w-4 h-4" /> {t('arch.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('arch.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-cyber-400 to-vortex-400 bg-clip-text text-transparent">{t('arch.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('arch.subtitle')}</p>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-20">
          <motion.div initial={{ opacity: 0, x: -40 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.6, delay: 0.2 }} className="space-y-3">
            {layers.map((layer, i) => (
              <motion.div key={layer.title} initial={{ opacity: 0, x: -30 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.4, delay: 0.3 + i * 0.1 }}>
                <div className="glass rounded-xl p-4 card-hover">
                  <div className="flex items-center gap-3 mb-2">
                    <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${layer.color} flex items-center justify-center text-white shrink-0`}>{layer.icon}</div>
                    <h4 className={`text-sm font-semibold ${isDark ? 'text-white/90' : 'text-gray-800'}`}>{layer.title}</h4>
                  </div>
                  <div className="flex flex-wrap gap-2 ms-11">
                    {layer.items.map((item) => (
                      <span key={item} className={`px-2.5 py-1 rounded-md text-xs font-mono ${isDark ? 'bg-white/5 text-white/50' : 'bg-gray-100 text-gray-500'}`}>{item}</span>
                    ))}
                  </div>
                </div>
                {i < layers.length - 1 && <div className="flex justify-center py-1"><div className={`w-px h-4 ${isDark ? 'bg-white/20' : 'bg-gray-300'}`} /></div>}
              </motion.div>
            ))}
          </motion.div>

          <motion.div initial={{ opacity: 0, x: 40 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.6, delay: 0.4 }}>
            <div className="glass rounded-2xl p-8 h-full">
              <h3 className={`text-xl font-bold mb-6 flex items-center gap-2 ${isDark ? 'text-white' : 'text-gray-900'}`}>
                <Cpu className="w-5 h-5 text-vortex-500" /> {t('arch.techStack')}
              </h3>
              <div className="space-y-4">
                {techStack.map((item, i) => (
                  <motion.div key={item.label} initial={{ opacity: 0, x: 20 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.3, delay: 0.5 + i * 0.05 }} className="flex items-start gap-4 group">
                    <div className="w-24 shrink-0"><span className="text-xs font-semibold text-vortex-500 uppercase tracking-wider">{item.label}</span></div>
                    <div className={`flex-1 py-2 border-b ${isDark ? 'border-white/5 group-hover:border-vortex-500/20' : 'border-gray-100 group-hover:border-vortex-200'} transition-colors`}>
                      <span className="text-sm text-themed-muted font-mono">{item.tech}</span>
                    </div>
                  </motion.div>
                ))}
              </div>
              <div className={`mt-8 pt-6 border-t ${isDark ? 'border-white/5' : 'border-gray-100'}`}>
                <h4 className="text-sm font-semibold text-themed-muted mb-4">{t('arch.dataFlow')}</h4>
                <div className="flex items-center justify-between text-xs font-mono text-themed-muted">
                  <span className="px-3 py-1.5 rounded-lg bg-cyan-500/10 text-cyan-500 border border-cyan-500/20">Client</span>
                  <span className="text-themed-faint">→</span>
                  <span className="px-3 py-1.5 rounded-lg bg-vortex-500/10 text-vortex-500 border border-vortex-500/20">Caddy</span>
                  <span className="text-themed-faint">→</span>
                  <span className="px-3 py-1.5 rounded-lg bg-amber-500/10 text-amber-500 border border-amber-500/20">API</span>
                  <span className="text-themed-faint">→</span>
                  <span className="px-3 py-1.5 rounded-lg bg-red-500/10 text-red-500 border border-red-500/20">Node</span>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
