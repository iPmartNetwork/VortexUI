import { motion, useInView } from 'framer-motion';
import { useRef, useState } from 'react';
import { Network, Check, X, Star } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

const protocols = [
  { name: 'VLESS', core: 'Both', inbound: true, outbound: true, transport: 'TCP, WS, gRPC, HTTPUpgrade, xHTTP, mKCP', security: 'None, TLS, REALITY', recommended: true },
  { name: 'VMess', core: 'Both', inbound: true, outbound: true, transport: 'TCP, WS, gRPC, HTTPUpgrade, mKCP', security: 'None, TLS' },
  { name: 'Trojan', core: 'Both', inbound: true, outbound: true, transport: 'TCP, WS, gRPC, mKCP', security: 'TLS, REALITY' },
  { name: 'Shadowsocks', core: 'Both', inbound: true, outbound: true, transport: 'TCP (+ SS-2022)', security: 'None' },
  { name: 'Hysteria2', core: 'sing-box', inbound: true, outbound: true, transport: 'UDP (QUIC)', security: 'TLS', recommended: true },
  { name: 'TUIC', core: 'sing-box', inbound: true, outbound: true, transport: 'UDP (QUIC)', security: 'TLS' },
  { name: 'WireGuard', core: 'sing-box', inbound: true, outbound: true, transport: 'UDP', security: 'Native' },
  { name: 'Hysteria v1', core: 'sing-box', inbound: true, outbound: false, transport: 'UDP', security: 'TLS' },
  { name: 'ShadowTLS', core: 'sing-box', inbound: true, outbound: true, transport: 'TCP', security: 'TLS' },
  { name: 'AnyTLS', core: 'sing-box', inbound: true, outbound: false, transport: 'TCP', security: 'TLS' },
  { name: 'Naive', core: 'sing-box', inbound: true, outbound: false, transport: '—', security: 'TLS (mandatory)' },
  { name: 'SOCKS', core: 'Both', inbound: true, outbound: true, transport: 'Raw TCP', security: 'Plaintext' },
  { name: 'HTTP', core: 'Both', inbound: true, outbound: true, transport: 'Raw TCP', security: 'Plaintext' },
  { name: 'Dokodemo', core: 'Xray', inbound: true, outbound: false, transport: 'Raw TCP/UDP', security: 'Plaintext' },
];

const subFormats = [
  { format: 'base64', client: 'v2rayNG / V2RayN', icon: '📦' },
  { format: 'clash', client: 'Clash Meta', icon: '⚔️' },
  { format: 'singbox', client: 'sing-box', icon: '📱' },
  { format: 'xray', client: 'Raw Xray', icon: '⚡' },
  { format: 'outline', client: 'Outline', icon: '🔑' },
  { format: 'links', client: 'One per line', icon: '🔗' },
];

export default function ProtocolsSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const [filter, setFilter] = useState<'all' | 'Both' | 'Xray' | 'sing-box'>('all');
  const { isDark } = useTheme();
  const { t } = useLang();

  const filtered = filter === 'all' ? protocols : protocols.filter(p => p.core === filter || p.core === 'Both');

  return (
    <section id="protocols" className="relative py-32">
      <div className="absolute inset-0 grid-bg opacity-50" />
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className={`inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm mb-6 ${isDark ? 'text-vortex-400' : 'text-vortex-600'}`}>
            <Network className="w-4 h-4" /> {t('proto.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('proto.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-vortex-400 to-cyber-400 bg-clip-text text-transparent">{t('proto.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('proto.subtitle')}</p>
        </motion.div>

        <div className="flex justify-center gap-2 mb-8">
          {(['all', 'Both', 'Xray', 'sing-box'] as const).map((f) => (
            <button key={f} onClick={() => setFilter(f)}
              className={`px-4 py-2 rounded-xl text-sm font-medium transition-all ${
                filter === f
                  ? 'bg-vortex-500/20 text-vortex-500 border border-vortex-500/30'
                  : isDark ? 'text-white/40 hover:text-white/70 glass' : 'text-gray-400 hover:text-gray-700 glass'
              }`}>
              {f === 'all' ? t('proto.filter.all') : f === 'Both' ? t('proto.filter.dual') : f}
            </button>
          ))}
        </div>

        <motion.div initial={{ opacity: 0, y: 20 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6, delay: 0.2 }} className="glass rounded-2xl overflow-hidden mb-16">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-themed">
                  <th className={`text-start text-xs font-semibold uppercase tracking-wider px-6 py-4 ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.protocol')}</th>
                  <th className={`text-start text-xs font-semibold uppercase tracking-wider px-6 py-4 ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.core')}</th>
                  <th className={`text-center text-xs font-semibold uppercase tracking-wider px-4 py-4 ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.in')}</th>
                  <th className={`text-center text-xs font-semibold uppercase tracking-wider px-4 py-4 ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.out')}</th>
                  <th className={`text-start text-xs font-semibold uppercase tracking-wider px-6 py-4 hidden md:table-cell ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.transport')}</th>
                  <th className={`text-start text-xs font-semibold uppercase tracking-wider px-6 py-4 hidden lg:table-cell ${isDark ? 'text-white/40' : 'text-gray-400'}`}>{t('proto.table.security')}</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map((p, i) => (
                  <motion.tr key={p.name} initial={{ opacity: 0, x: -10 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.3, delay: 0.3 + i * 0.03 }}
                    className={`border-b border-themed hover:bg-[var(--bg-card-hover)] transition-colors`}>
                    <td className="px-6 py-3.5">
                      <div className="flex items-center gap-2">
                        <span className={`text-sm font-medium ${isDark ? 'text-white/90' : 'text-gray-800'}`}>{p.name}</span>
                        {p.recommended && (<span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-500 text-[10px] font-semibold"><Star className="w-2.5 h-2.5" />REC</span>)}
                      </div>
                    </td>
                    <td className="px-6 py-3.5">
                      <span className={`text-xs font-mono px-2 py-1 rounded-md ${p.core === 'Both' ? 'bg-vortex-500/10 text-vortex-500' : p.core === 'Xray' ? 'bg-cyan-500/10 text-cyan-500' : 'bg-amber-500/10 text-amber-500'}`}>{p.core}</span>
                    </td>
                    <td className="px-4 py-3.5 text-center">{p.inbound ? <Check className="w-4 h-4 text-green-500 mx-auto" /> : <X className="w-4 h-4 text-themed-faint mx-auto" />}</td>
                    <td className="px-4 py-3.5 text-center">{p.outbound ? <Check className="w-4 h-4 text-green-500 mx-auto" /> : <X className="w-4 h-4 text-themed-faint mx-auto" />}</td>
                    <td className="px-6 py-3.5 hidden md:table-cell"><span className="text-xs text-themed-muted font-mono">{p.transport}</span></td>
                    <td className="px-6 py-3.5 hidden lg:table-cell"><span className="text-xs text-themed-muted font-mono">{p.security}</span></td>
                  </motion.tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

        <motion.div initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6, delay: 0.4 }}>
          <h3 className={`text-2xl font-bold text-center mb-8 ${isDark ? 'text-white/80' : 'text-gray-700'}`}>{t('proto.subFormats')}</h3>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            {subFormats.map((f, i) => (
              <motion.div key={f.format} initial={{ opacity: 0, scale: 0.9 }} animate={isInView ? { opacity: 1, scale: 1 } : {}} transition={{ duration: 0.3, delay: 0.5 + i * 0.05 }}
                className="glass rounded-xl p-4 text-center card-hover">
                <div className="text-2xl mb-2">{f.icon}</div>
                <div className="text-sm font-mono text-vortex-500 font-semibold">{f.format}</div>
                <div className="text-xs text-themed-faint mt-1">{f.client}</div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>
    </section>
  );
}
