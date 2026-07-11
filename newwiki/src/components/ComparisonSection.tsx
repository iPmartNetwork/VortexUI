import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import { Trophy, Check, X, Minus } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

type CellValue = boolean | string;
interface CompRow { feature: string; vortex: CellValue; threexui: CellValue; marzban: CellValue; hiddify: CellValue; }

const comparisons: CompRow[] = [
  { feature: 'Dual core (Xray + sing-box)', vortex: true, threexui: false, marzban: false, hiddify: true },
  { feature: 'User-centric model', vortex: true, threexui: false, marzban: true, hiddify: true },
  { feature: 'Push delta traffic', vortex: true, threexui: 'polling', marzban: 'polling', hiddify: 'polling' },
  { feature: 'Node auto-migration', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Load balancer (4 strategies)', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Reality Scanner', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'TLS Tricks (ISP profiles)', vortex: true, threexui: false, marzban: false, hiddify: 'partial' },
  { feature: 'Probing protection', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'JA3 fingerprint validation', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Self-service portal', vortex: true, threexui: false, marzban: false, hiddify: true },
  { feature: 'Per-reseller shop', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Federation', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Family groups', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Smart Quota', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'CDN/Relay chains', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Command palette (Ctrl+K)', vortex: true, threexui: false, marzban: false, hiddify: false },
  { feature: 'Backend', vortex: 'Go', threexui: 'Go', marzban: 'Python', hiddify: 'Python' },
  { feature: 'Database', vortex: 'PG+TS', threexui: 'SQLite', marzban: 'SQLite', hiddify: 'SQLite' },
];

function CellRenderer({ value, isDark }: { value: CellValue; isDark: boolean }) {
  if (value === true) return <Check className="w-4.5 h-4.5 text-green-500 mx-auto" />;
  if (value === false) return <X className={`w-4.5 h-4.5 mx-auto ${isDark ? 'text-white/15' : 'text-gray-200'}`} />;
  if (value === 'partial') return <Minus className="w-4.5 h-4.5 text-amber-400/60 mx-auto" />;
  return <span className="text-xs font-mono text-themed-muted">{value}</span>;
}

export default function ComparisonSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  const scores = [
    { name: 'VortexUI', score: 18, total: 18, color: 'from-vortex-500 to-cyber-400' },
    { name: '3x-ui', score: 2, total: 18, color: 'from-gray-400 to-gray-500' },
    { name: 'Marzban', score: 3, total: 18, color: 'from-gray-400 to-gray-500' },
    { name: 'Hiddify', score: 5, total: 18, color: 'from-gray-400 to-gray-500' },
  ];

  return (
    <section id="comparison" className="relative py-32">
      <div className="absolute inset-0 grid-bg opacity-30" />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm text-amber-500 mb-6">
            <Trophy className="w-4 h-4" /> {t('cmp.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('cmp.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-amber-400 to-orange-500 bg-clip-text text-transparent">{t('cmp.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('cmp.subtitle')}</p>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6, delay: 0.2 }} className="glass rounded-2xl overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-themed">
                  <th className={`text-start text-xs font-semibold uppercase tracking-wider px-6 py-4 min-w-[200px] ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('cmp.feature')}</th>
                  <th className="text-center px-4 py-4 min-w-[100px]">
                    <span className="text-sm font-bold bg-gradient-to-r from-vortex-500 to-cyber-400 bg-clip-text text-transparent">VortexUI</span>
                  </th>
                  <th className={`text-center text-xs font-semibold uppercase px-4 py-4 min-w-[80px] ${isDark ? 'text-white/40' : 'text-gray-400'}`}>3x-ui</th>
                  <th className={`text-center text-xs font-semibold uppercase px-4 py-4 min-w-[80px] ${isDark ? 'text-white/40' : 'text-gray-400'}`}>Marzban</th>
                  <th className={`text-center text-xs font-semibold uppercase px-4 py-4 min-w-[80px] ${isDark ? 'text-white/40' : 'text-gray-400'}`}>Hiddify</th>
                </tr>
              </thead>
              <tbody>
                {comparisons.map((row, i) => (
                  <tr key={row.feature} className={`border-b border-themed hover:bg-[var(--bg-card-hover)] transition-colors ${i % 2 === 0 ? 'bg-[var(--bg-card)]' : ''}`}>
                    <td className={`px-6 py-3 text-sm ${isDark ? 'text-white/70' : 'text-gray-600'}`}>{row.feature}</td>
                    <td className="px-4 py-3 text-center"><div className="flex justify-center"><div className={row.vortex === true ? 'bg-green-500/10 rounded-full p-1' : ''}><CellRenderer value={row.vortex} isDark={isDark} /></div></div></td>
                    <td className="px-4 py-3 text-center"><CellRenderer value={row.threexui} isDark={isDark} /></td>
                    <td className="px-4 py-3 text-center"><CellRenderer value={row.marzban} isDark={isDark} /></td>
                    <td className="px-4 py-3 text-center"><CellRenderer value={row.hiddify} isDark={isDark} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6, delay: 0.4 }} className="mt-8 flex flex-wrap justify-center gap-6">
          {scores.map((p) => (
            <div key={p.name} className="glass rounded-xl px-6 py-4 text-center min-w-[140px]">
              <div className={`text-2xl font-bold bg-gradient-to-r ${p.color} bg-clip-text text-transparent`}>{p.score}/{p.total}</div>
              <div className="text-xs text-themed-muted mt-1">{p.name}</div>
              <div className={`mt-2 h-1.5 rounded-full overflow-hidden ${isDark ? 'bg-white/5' : 'bg-gray-100'}`}>
                <div className={`h-full rounded-full bg-gradient-to-r ${p.color}`} style={{ width: `${(p.score / p.total) * 100}%` }} />
              </div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
