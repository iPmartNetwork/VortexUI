import { motion, useInView } from 'framer-motion';
import { useRef, useState } from 'react';
import { Terminal, Copy, Check, Download, Server, Container, Code2 } from 'lucide-react';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

type Tab = 'one-line' | 'docker' | 'native' | 'node';

const codeBlocks: Record<Tab, { cmd: string; steps: string[] }> = {
  'one-line': {
    cmd: 'bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)',
    steps: ['Detects OS and architecture', 'Installs PostgreSQL, Redis, Caddy', 'Downloads and builds VortexUI', 'Creates sudo admin (interactive)', 'Configures HTTPS via Caddy'],
  },
  docker: {
    cmd: 'git clone https://github.com/iPmartNetwork/VortexUI.git\ncd VortexUI/deploy\ncp ../.env.example .env\n# Edit .env with your settings\ndocker compose up -d',
    steps: ['Includes panel, PostgreSQL + TimescaleDB, Redis, Caddy', 'Edit .env for domain, admin, JWT secret', 'Auto-configures HTTPS'],
  },
  native: {
    cmd: 'sudo snap install go --classic\nsudo apt update && sudo apt install -y postgresql redis-server\ngit clone https://github.com/iPmartNetwork/VortexUI.git\ncd VortexUI\ngo build -o vortexui ./cmd/panel\n./vortexui migrate\n./vortexui admin create --username admin --password your-password --sudo\n./vortexui serve',
    steps: ['Requires Go 1.26+, PostgreSQL, Redis', 'Manual build and migration', 'Full control over configuration'],
  },
  node: {
    cmd: 'bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install-node.sh)',
    steps: ['Run on remote server', 'Provide panel address and enrollment token', 'Auto-registers, exchanges mTLS certificates'],
  },
};

const requirements = [
  { label: 'OS', min: 'Ubuntu 20.04', rec: 'Ubuntu 22.04+' },
  { label: 'RAM', min: '1 GB', rec: '2 GB+' },
  { label: 'Disk', min: '10 GB', rec: '20 GB+' },
  { label: 'CPU', min: '1 vCPU', rec: '2+ vCPU' },
];

export default function InstallSection() {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: '-80px' });
  const [activeTab, setActiveTab] = useState<Tab>('one-line');
  const [copied, setCopied] = useState(false);
  const { isDark } = useTheme();
  const { t } = useLang();

  const tabs: { id: Tab; label: string; icon: React.ReactNode }[] = [
    { id: 'one-line', label: t('inst.tab.quick'), icon: <Terminal className="w-4 h-4" /> },
    { id: 'docker', label: t('inst.tab.docker'), icon: <Container className="w-4 h-4" /> },
    { id: 'native', label: t('inst.tab.native'), icon: <Code2 className="w-4 h-4" /> },
    { id: 'node', label: t('inst.tab.node'), icon: <Server className="w-4 h-4" /> },
  ];

  const handleCopy = () => { navigator.clipboard.writeText(codeBlocks[activeTab].cmd); setCopied(true); setTimeout(() => setCopied(false), 2000); };

  return (
    <section id="install" className="relative py-32">
      <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <motion.div ref={ref} initial={{ opacity: 0, y: 30 }} animate={isInView ? { opacity: 1, y: 0 } : {}} transition={{ duration: 0.6 }} className="text-center mb-16">
          <div className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full glass text-sm text-green-500 mb-6">
            <Download className="w-4 h-4" /> {t('inst.badge')}
          </div>
          <h2 className="text-4xl sm:text-5xl lg:text-6xl font-bold mb-6">
            <span className={`${isDark ? 'bg-gradient-to-b from-white to-white/60' : 'bg-gradient-to-b from-gray-900 to-gray-500'} bg-clip-text text-transparent`}>{t('inst.title.line1')}</span>{' '}
            <span className="bg-gradient-to-r from-green-400 to-cyber-400 bg-clip-text text-transparent">{t('inst.title.line2')}</span>
          </h2>
          <p className="text-lg text-themed-muted max-w-2xl mx-auto">{t('inst.subtitle')}</p>
        </motion.div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          <motion.div initial={{ opacity: 0, x: -30 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.6, delay: 0.2 }} className="lg:col-span-2">
            <div className="flex flex-wrap gap-2 mb-4">
              {tabs.map((tab) => (
                <button key={tab.id} onClick={() => setActiveTab(tab.id)}
                  className={`flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium transition-all ${
                    activeTab === tab.id ? 'bg-vortex-500/20 text-vortex-500 border border-vortex-500/30'
                    : isDark ? 'text-white/40 hover:text-white/70 glass' : 'text-gray-400 hover:text-gray-700 glass'
                  }`}>{tab.icon}{tab.label}</button>
              ))}
            </div>
            <div className="relative">
              <div className="absolute -inset-1 bg-gradient-to-r from-green-500/10 via-vortex-500/10 to-cyber-400/10 rounded-2xl blur-xl" />
              <div className="relative glass-strong rounded-2xl overflow-hidden">
                <div className="flex items-center justify-between px-5 py-3 border-b border-themed">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full bg-red-500/80" /><div className="w-3 h-3 rounded-full bg-yellow-500/80" /><div className="w-3 h-3 rounded-full bg-green-500/80" />
                  </div>
                  <span className="text-xs text-themed-faint font-mono">{activeTab}</span>
                  <button onClick={handleCopy} className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs text-themed-muted hover:text-themed transition-all">
                    {copied ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
                    {copied ? t('inst.copied') : t('inst.copy')}
                  </button>
                </div>
                <div className="p-6 font-mono text-sm overflow-x-auto bg-[var(--code-bg)] text-white">
                  {codeBlocks[activeTab].cmd.split('\n').map((line, i) => (
                    <div key={i} className="flex items-start gap-3">
                      <span className="text-white/20 select-none shrink-0 w-4 text-right">{i + 1}</span>
                      {line.startsWith('#') ? <span className="text-white/30">{line}</span> : <span className="text-green-400/90">{line}</span>}
                    </div>
                  ))}
                </div>
                <div className={`px-6 pb-6 pt-4 border-t border-themed ${isDark ? '' : 'bg-white/50'}`}>
                  <div className="text-xs text-themed-faint mb-3 font-semibold uppercase tracking-wider">{t('inst.whatHappens')}</div>
                  <div className="space-y-2">
                    {codeBlocks[activeTab].steps.map((step, i) => (
                      <div key={i} className="flex items-center gap-2 text-xs text-themed-muted"><Check className="w-3.5 h-3.5 text-green-500/60 shrink-0" />{step}</div>
                    ))}
                  </div>
                </div>
              </div>
            </div>
          </motion.div>

          <motion.div initial={{ opacity: 0, x: 30 }} animate={isInView ? { opacity: 1, x: 0 } : {}} transition={{ duration: 0.6, delay: 0.4 }} className="space-y-6">
            <div className="glass rounded-2xl p-6">
              <h3 className={`text-lg font-semibold mb-4 flex items-center gap-2 ${isDark ? 'text-white' : 'text-gray-900'}`}>
                <Server className="w-5 h-5 text-vortex-500" /> {t('inst.requirements')}
              </h3>
              <div className="space-y-4">
                {requirements.map((req) => (
                  <div key={req.label}>
                    <div className="text-xs font-semibold text-themed-muted uppercase tracking-wider mb-1">{req.label}</div>
                    <div className="text-sm text-themed-muted">Min: <span className="text-themed-secondary font-mono">{req.min}</span></div>
                    <div className="text-sm text-themed-muted">Rec: <span className="text-green-500 font-mono">{req.rec}</span></div>
                  </div>
                ))}
              </div>
            </div>
            <div className="glass rounded-2xl p-6">
              <h3 className={`text-lg font-semibold mb-4 ${isDark ? 'text-white' : 'text-gray-900'}`}>{t('inst.postInstall')}</h3>
              <div className="space-y-3 text-sm text-themed-muted">
                {['https://your-domain.com', 'Login with admin', 'Add a node', 'Create inbound', 'vortexui doctor'].map((s, i) => (
                  <div key={i} className="flex items-start gap-2"><span className="text-green-500 mt-0.5">{i + 1}.</span><span>{s}</span></div>
                ))}
              </div>
            </div>
            <div className="glass rounded-2xl p-6 border border-vortex-500/10">
              <div className="text-sm text-themed-muted mb-2">{t('inst.healthCheck')}</div>
              <code className="text-xs text-cyber-500 font-mono">GET /api/health → 200 OK</code>
            </div>
          </motion.div>
        </div>
      </div>
    </section>
  );
}
