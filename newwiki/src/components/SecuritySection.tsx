import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import { Shield, Lock, Fingerprint, Eye, Scan, Radio, FileWarning, Globe, UserX, Key, ShieldCheck, Waves } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

export default function SecuritySection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  const features = [
    { icon: <Scan className="w-7 h-7" />, title: t('sec.realityScanner.title'), desc: t('sec.realityScanner.desc'), color: 'text-cyan-500', border: 'border-cyan-500/20 hover:border-cyan-500/40', bg: 'bg-cyan-500/5' },
    { icon: <Fingerprint className="w-7 h-7" />, title: t('sec.tlsTricks.title'), desc: t('sec.tlsTricks.desc'), color: 'text-vortex-500', border: 'border-vortex-500/20 hover:border-vortex-500/40', bg: 'bg-vortex-500/5' },
    { icon: <Shield className="w-7 h-7" />, title: t('sec.probing.title'), desc: t('sec.probing.desc'), color: 'text-red-500', border: 'border-red-500/20 hover:border-red-500/40', bg: 'bg-red-500/5' },
    { icon: <Eye className="w-7 h-7" />, title: t('sec.ja3.title'), desc: t('sec.ja3.desc'), color: 'text-amber-500', border: 'border-amber-500/20 hover:border-amber-500/40', bg: 'bg-amber-500/5' },
    { icon: <FileWarning className="w-7 h-7" />, title: t('sec.decoy.title'), desc: t('sec.decoy.desc'), color: 'text-green-500', border: 'border-green-500/20 hover:border-green-500/40', bg: 'bg-green-500/5' },
    { icon: <Radio className="w-7 h-7" />, title: t('sec.doh.title'), desc: t('sec.doh.desc'), color: 'text-blue-500', border: 'border-blue-500/20 hover:border-blue-500/40', bg: 'bg-blue-500/5' },
    { icon: <Waves className="w-7 h-7" />, title: t('sec.warp.title'), desc: t('sec.warp.desc'), color: 'text-orange-500', border: 'border-orange-500/20 hover:border-orange-500/40', bg: 'bg-orange-500/5' },
    { icon: <Globe className="w-7 h-7" />, title: t('sec.cleanIp.title'), desc: t('sec.cleanIp.desc'), color: 'text-teal-500', border: 'border-teal-500/20 hover:border-teal-500/40', bg: 'bg-teal-500/5' },
    { icon: <UserX className="w-7 h-7" />, title: t('sec.sharing.title'), desc: t('sec.sharing.desc'), color: 'text-pink-500', border: 'border-pink-500/20 hover:border-pink-500/40', bg: 'bg-pink-500/5' },
    { icon: <Lock className="w-7 h-7" />, title: t('sec.rbac.title'), desc: t('sec.rbac.desc'), color: 'text-purple-500', border: 'border-purple-500/20 hover:border-purple-500/40', bg: 'bg-purple-500/5' },
    { icon: <Key className="w-7 h-7" />, title: t('sec.mtls.title'), desc: t('sec.mtls.desc'), color: 'text-indigo-500', border: 'border-indigo-500/20 hover:border-indigo-500/40', bg: 'bg-indigo-500/5' },
    { icon: <ShieldCheck className="w-7 h-7" />, title: t('sec.evasion.title'), desc: t('sec.evasion.desc'), color: 'text-emerald-500', border: 'border-emerald-500/20 hover:border-emerald-500/40', bg: 'bg-emerald-500/5' },
  ];

  return (
    <section id="security" className="relative py-32 overflow-hidden">
      <div className={`absolute top-1/4 left-0 w-[500px] h-[500px] ${isDark ? 'bg-red-500/3' : 'bg-red-500/2'} rounded-full blur-3xl`} />
      <div className={`absolute bottom-1/4 right-0 w-[500px] h-[500px] ${isDark ? 'bg-vortex-500/3' : 'bg-vortex-500/2'} rounded-full blur-3xl`} />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm text-red-500 mb-6">
            <Shield className="w-4 h-4" /> {t('sec.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('sec.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-red-500 to-vortex-500 bg-clip-text text-transparent">{t('sec.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('sec.subtitle')}</p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {features.map((f, i) => (
            <motion.div key={f.title} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.4, delay: i * 0.05 }}
              className={`glass rounded-2xl p-6 border ${f.border} card-hover group`}>
              <div className={`w-14 h-14 rounded-xl ${f.bg} flex items-center justify-center mb-4 ${f.color} group-hover:scale-110 transition-transform`}>{f.icon}</div>
              <h3 className={`text-base font-semibold mb-2 ${isDark ? 'text-white' : 'text-gray-900'}`}>{f.title}</h3>
              <p className="text-sm text-themed-muted leading-relaxed">{f.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
