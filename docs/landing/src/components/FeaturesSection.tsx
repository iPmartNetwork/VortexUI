import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import {
  Shield, Globe, Users, Server, Activity, Palette,
  ShoppingCart, Lock, BarChart3, Layers, Wifi, Bell, Fingerprint,
  ArrowRightLeft, Brain, Network, Shuffle, Gauge
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

export default function FeaturesSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-100px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  const features = [
    { icon: <Server className="w-6 h-6" />, title: t('features.dualCore.title'), desc: t('features.dualCore.desc'), color: 'from-vortex-500 to-vortex-700' },
    { icon: <Shield className="w-6 h-6" />, title: t('features.antiCensor.title'), desc: t('features.antiCensor.desc'), color: 'from-red-500 to-rose-700' },
    { icon: <Globe className="w-6 h-6" />, title: t('features.nodeFleet.title'), desc: t('features.nodeFleet.desc'), color: 'from-cyan-500 to-cyan-700' },
    { icon: <Users className="w-6 h-6" />, title: t('features.userCentric.title'), desc: t('features.userCentric.desc'), color: 'from-green-500 to-emerald-700' },
    { icon: <ShoppingCart className="w-6 h-6" />, title: t('features.portal.title'), desc: t('features.portal.desc'), color: 'from-amber-500 to-orange-700' },
    { icon: <Lock className="w-6 h-6" />, title: t('features.reseller.title'), desc: t('features.reseller.desc'), color: 'from-purple-500 to-indigo-700' },
    { icon: <BarChart3 className="w-6 h-6" />, title: t('features.analytics.title'), desc: t('features.analytics.desc'), color: 'from-blue-500 to-blue-700' },
    { icon: <Palette className="w-6 h-6" />, title: t('features.veltrix.title'), desc: t('features.veltrix.desc'), color: 'from-pink-500 to-fuchsia-700' },
    { icon: <Layers className="w-6 h-6" />, title: t('features.cdn.title'), desc: t('features.cdn.desc'), color: 'from-teal-500 to-teal-700' },
    { icon: <Wifi className="w-6 h-6" />, title: t('features.federation.title'), desc: t('features.federation.desc'), color: 'from-sky-500 to-sky-700' },
    { icon: <Bell className="w-6 h-6" />, title: t('features.notifications.title'), desc: t('features.notifications.desc'), color: 'from-yellow-500 to-yellow-700' },
    { icon: <Fingerprint className="w-6 h-6" />, title: t('features.security.title'), desc: t('features.security.desc'), color: 'from-red-600 to-rose-800' },
    { icon: <ArrowRightLeft className="w-6 h-6" />, title: t('features.autoSwitch.title'), desc: t('features.autoSwitch.desc'), color: 'from-emerald-500 to-emerald-700' },
    { icon: <Brain className="w-6 h-6" />, title: t('features.smartConfig.title'), desc: t('features.smartConfig.desc'), color: 'from-violet-500 to-violet-700' },
    { icon: <Network className="w-6 h-6" />, title: t('features.multiPath.title'), desc: t('features.multiPath.desc'), color: 'from-indigo-500 to-indigo-700' },
    { icon: <Shuffle className="w-6 h-6" />, title: t('features.dynamicSNI.title'), desc: t('features.dynamicSNI.desc'), color: 'from-rose-500 to-rose-700' },
    { icon: <Gauge className="w-6 h-6" />, title: t('features.qualityScore.title'), desc: t('features.qualityScore.desc'), color: 'from-orange-500 to-orange-700' },
  ];

  return (
    <section id="features" className="relative py-32 grid-bg">
      <div className={`absolute top-0 left-1/2 -translate-x-1/2 w-[800px] h-[400px] ${isDark ? 'bg-vortex-500/5' : 'bg-vortex-500/3'} rounded-full blur-3xl`} />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm ${isDark ? 'text-vortex-400' : 'text-vortex-600'} mb-6`}>
            <Activity className="w-4 h-4" /> {t('features.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>
              {t('features.title.line1')}
            </span><br />
            <span className="bg-gradient-to-r from-vortex-400 to-cyber-400 bg-clip-text text-transparent">{t('features.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('features.subtitle')}</p>
        </motion.div>

        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
          {features.map((f, i) => (
            <motion.div key={f.title} initial={{ opacity: 0, y: 40 }} animate={isInView ? { opacity: 1, y: 0 } : {}}
              transition={{ duration: 0.5, delay: i * 0.06 }} className="group relative">
              <div className={`absolute -inset-0.5 rounded-2xl bg-gradient-to-r ${f.color} opacity-0 group-hover:opacity-10 blur-xl transition-opacity duration-500`} />
              <div className="relative h-full glass rounded-2xl p-6 card-hover">
                <div className={`w-12 h-12 rounded-xl bg-gradient-to-br ${f.color} flex items-center justify-center mb-4 shadow-lg`}>
                  <div className="text-white">{f.icon}</div>
                </div>
                <h3 className={`text-lg font-semibold mb-2 ${isDark ? 'text-white' : 'text-gray-900'}`}>{f.title}</h3>
                <p className="text-sm text-themed-muted leading-relaxed">{f.desc}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}
