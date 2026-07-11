import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import {
  LayoutDashboard, Users, Server, Activity, Globe, TrendingUp,
  Zap, ArrowUp, ArrowDown, Wifi, Clock
} from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

function MiniGauge({ label, value, color }: { label: string; value: number; color: string }) {
  const pct = value * 0.942;
  return (
    <div className="text-center">
      <div className="relative w-16 h-16 mx-auto mb-2">
        <svg className="w-full h-full transform -rotate-90" viewBox="0 0 36 36">
          <circle cx="18" cy="18" r="15" fill="none" stroke="rgba(255,255,255,0.05)" strokeWidth="3" />
          <circle cx="18" cy="18" r="15" fill="none" stroke={color} strokeWidth="3" strokeDasharray={`${pct} 94.2`} strokeLinecap="round" />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center text-xs font-bold text-white/80">{value}%</div>
      </div>
      <div className="text-[10px] text-white/40">{label}</div>
    </div>
  );
}

function StatCard({ icon, label, value, change, trend }: { icon: React.ReactNode; label: string; value: string; change: string; trend: 'up' | 'down' }) {
  return (
    <div className="bg-white/[0.02] rounded-xl p-3 border border-white/5">
      <div className="flex items-center justify-between mb-2">
        <div className="text-white/30">{icon}</div>
        <div className={`flex items-center gap-0.5 text-[10px] ${trend === 'up' ? 'text-green-400' : 'text-red-400'}`}>
          {trend === 'up' ? <ArrowUp className="w-2.5 h-2.5" /> : <ArrowDown className="w-2.5 h-2.5" />}{change}
        </div>
      </div>
      <div className="text-lg font-bold text-white/90">{value}</div>
      <div className="text-[10px] text-white/30">{label}</div>
    </div>
  );
}

function TrafficChart() {
  const bars = [35, 52, 45, 70, 85, 62, 48, 90, 75, 55, 68, 42, 78, 95, 60, 45, 82, 58, 72, 88, 50, 65, 40, 73];
  return (
    <div className="flex items-end gap-[3px] h-20">
      {bars.map((h, i) => (
        <div key={i} className="flex-1 rounded-t-sm bg-gradient-to-t from-vortex-500/60 to-cyber-400/40 hover:from-vortex-400 hover:to-cyber-400 transition-all" style={{ height: `${h}%` }} />
      ))}
    </div>
  );
}

export default function DashboardPreview() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const { isDark } = useTheme();
  const { t } = useLang();

  const badges = [
    t('dashboard.commandPalette'), t('dashboard.dragDrop'), t('dashboard.darkLight'),
    t('dashboard.languages'), t('dashboard.mobile'), t('dashboard.sse'),
  ];

  return (
    <section className="relative py-32 overflow-hidden">
      <div className="absolute inset-0 grid-bg opacity-30" />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm mb-6 ${isDark ? 'text-vortex-400' : 'text-vortex-600'}`}>
            <LayoutDashboard className="w-4 h-4" /> {t('dashboard.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('dashboard.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-vortex-400 to-cyber-400 bg-clip-text text-transparent">{t('dashboard.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('dashboard.subtitle')}</p>
        </motion.div>

        {/* Dashboard Mockup — always dark themed to look like the actual panel */}
        <motion.div initial={{ opacity: 0, y: 40, scale: 0.95 }} animate={isInView ? { opacity: 1, y: 0, scale: 1 } : {}} transition={{ duration: 0.8, delay: 0.2 }}
          className="relative max-w-5xl mx-auto">
          <div className="absolute -inset-4 bg-gradient-to-r from-vortex-500/10 via-cyber-400/10 to-vortex-500/10 rounded-3xl blur-2xl" />
          <div className="relative bg-[#0a0a1a] rounded-2xl overflow-hidden border border-white/10 text-white">
            <div className="flex items-center justify-between px-4 py-3 border-b border-white/5 bg-white/[0.02]">
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full bg-red-500/80" /><div className="w-3 h-3 rounded-full bg-yellow-500/80" /><div className="w-3 h-3 rounded-full bg-green-500/80" />
              </div>
              <div className="text-xs text-white/20 font-mono flex items-center gap-2">
                <Zap className="w-3 h-3 text-vortex-400" /> VortexUI Dashboard
              </div>
              <div className="text-[10px] text-white/15 font-mono">Ctrl+K</div>
            </div>
            <div className="flex">
              <div className="hidden sm:block w-14 shrink-0 border-r border-white/5 py-4 bg-white/[0.01]">
                {[LayoutDashboard, Users, Server, Globe, Activity, Wifi, Clock].map((Icon, i) => (
                  <div key={i} className={`w-10 h-10 mx-auto mb-1 rounded-lg flex items-center justify-center ${i === 0 ? 'bg-vortex-500/20 text-vortex-400' : 'text-white/20'}`}>
                    <Icon className="w-4 h-4" />
                  </div>
                ))}
              </div>
              <div className="flex-1 p-4 space-y-4">
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  <StatCard icon={<Users className="w-4 h-4" />} label={t('dashboard.totalUsers')} value="2,847" change="+12%" trend="up" />
                  <StatCard icon={<Server className="w-4 h-4" />} label={t('dashboard.onlineNodes')} value="18/20" change="-1" trend="down" />
                  <StatCard icon={<TrendingUp className="w-4 h-4" />} label={t('dashboard.todayTraffic')} value="1.8 TB" change="+25%" trend="up" />
                  <StatCard icon={<Activity className="w-4 h-4" />} label={t('dashboard.activeConns')} value="1,205" change="+8%" trend="up" />
                </div>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <div className="sm:col-span-2 bg-white/[0.02] rounded-xl p-4 border border-white/5">
                    <div className="flex items-center justify-between mb-3">
                      <div className="text-xs font-semibold text-white/50">{t('dashboard.traffic24')}</div>
                      <div className="flex gap-2">{['24h', '7d', '30d'].map((tt, i) => (<span key={tt} className={`text-[9px] px-1.5 py-0.5 rounded ${i === 0 ? 'bg-vortex-500/20 text-vortex-400' : 'text-white/20'}`}>{tt}</span>))}</div>
                    </div>
                    <TrafficChart />
                  </div>
                  <div className="bg-white/[0.02] rounded-xl p-4 border border-white/5">
                    <div className="text-xs font-semibold text-white/50 mb-3">{t('dashboard.systemHealth')}</div>
                    <div className="grid grid-cols-2 gap-3">
                      <MiniGauge label="CPU" value={42} color="#6c5ce7" />
                      <MiniGauge label="RAM" value={67} color="#00f5ff" />
                      <MiniGauge label="Disk" value={35} color="#39ff14" />
                      <MiniGauge label="Net" value={78} color="#ff006e" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.5, delay: 0.6 }}
          className="mt-10 flex flex-wrap justify-center gap-3">
          {badges.map((b) => (
            <span key={b} className={`px-3 py-1.5 rounded-full glass text-xs ${isDark ? 'text-white/50' : 'text-gray-500'} border border-themed`}>{b}</span>
          ))}
        </motion.div>
      </div>
    </section>
  );
}
