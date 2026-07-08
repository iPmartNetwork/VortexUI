import { useEffect, useState, useRef } from 'react';
import { motion } from 'framer-motion';
import { ArrowRight, Terminal, Copy, Check, Sparkles, Shield, Globe, Zap } from 'lucide-react';
import { GithubIcon } from './icons';
import { useTheme } from '../contexts/ThemeContext';
import { useLang } from '../contexts/LangContext';

function ParticleField({ isDark }: { isDark: boolean }) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    let animationId: number;
    const particles: { x: number; y: number; vx: number; vy: number; size: number; opacity: number; color: string }[] = [];
    const resize = () => { canvas.width = window.innerWidth; canvas.height = window.innerHeight; };
    resize();
    window.addEventListener('resize', resize);
    const darkColors = ['rgba(108,92,231,', 'rgba(0,245,255,', 'rgba(255,0,110,'];
    const lightColors = ['rgba(108,92,231,', 'rgba(90,63,214,', 'rgba(76,50,181,'];
    const colors = isDark ? darkColors : lightColors;
    for (let i = 0; i < 60; i++) {
      particles.push({
        x: Math.random() * canvas.width, y: Math.random() * canvas.height,
        vx: (Math.random() - 0.5) * 0.3, vy: (Math.random() - 0.5) * 0.3,
        size: Math.random() * 2 + 0.5,
        opacity: isDark ? Math.random() * 0.5 + 0.1 : Math.random() * 0.3 + 0.05,
        color: colors[Math.floor(Math.random() * colors.length)],
      });
    }
    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
      particles.forEach((p) => {
        p.x += p.vx; p.y += p.vy;
        if (p.x < 0) p.x = canvas.width; if (p.x > canvas.width) p.x = 0;
        if (p.y < 0) p.y = canvas.height; if (p.y > canvas.height) p.y = 0;
        ctx.beginPath(); ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2);
        ctx.fillStyle = `${p.color}${p.opacity})`; ctx.fill();
      });
      for (let i = 0; i < particles.length; i++) {
        for (let j = i + 1; j < particles.length; j++) {
          const dx = particles[i].x - particles[j].x;
          const dy = particles[i].y - particles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 120) {
            ctx.beginPath(); ctx.moveTo(particles[i].x, particles[i].y); ctx.lineTo(particles[j].x, particles[j].y);
            ctx.strokeStyle = `rgba(108,92,231,${(isDark ? 0.06 : 0.03) * (1 - dist / 120)})`;
            ctx.lineWidth = 0.5; ctx.stroke();
          }
        }
      }
      animationId = requestAnimationFrame(animate);
    };
    animate();
    return () => { cancelAnimationFrame(animationId); window.removeEventListener('resize', resize); };
  }, [isDark]);
  return <canvas ref={canvasRef} className="absolute inset-0 z-0" />;
}

function TerminalDemo() {
  const [copied, setCopied] = useState(false);
  const { t } = useLang();
  const cmd = 'bash <(curl -Ls https://raw.githubusercontent.com/iPmartNetwork/VortexUI/master/install.sh)';
  const handleCopy = () => { navigator.clipboard.writeText(cmd); setCopied(true); setTimeout(() => setCopied(false), 2000); };

  return (
    <motion.div initial={{ opacity: 0, y: 30, scale: 0.95 }} animate={{ opacity: 1, y: 0, scale: 1 }} transition={{ duration: 0.8, delay: 0.6 }}
      className="relative max-w-2xl mx-auto">
      <div className="absolute -inset-1 bg-gradient-to-r from-vortex-500/20 via-cyber-400/20 to-vortex-500/20 rounded-2xl blur-xl animate-gradient" />
      <div className="relative glass-strong rounded-2xl overflow-hidden">
        <div className="flex items-center justify-between px-4 py-3 border-b border-themed">
          <div className="flex items-center gap-2">
            <div className="w-3 h-3 rounded-full bg-red-500/80" />
            <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
            <div className="w-3 h-3 rounded-full bg-green-500/80" />
          </div>
          <span className="text-xs text-themed-faint font-mono">bash — vortex-install</span>
          <button onClick={handleCopy} className="flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs text-themed-muted hover:text-themed transition-all">
            {copied ? <Check className="w-3.5 h-3.5 text-green-400" /> : <Copy className="w-3.5 h-3.5" />}
            {copied ? t('inst.copied') : t('inst.copy')}
          </button>
        </div>
        <div className="p-5 font-mono text-sm bg-[var(--code-bg)] text-white">
          <div className="flex items-start gap-2"><span className="text-green-400 shrink-0">$</span><span className="text-white/90 break-all">{cmd}</span></div>
          <div className="mt-3 text-white/40 text-xs"><div className="flex items-center gap-2"><span className="text-cyber-400">✓</span><span>{t('hero.terminal.comment')}</span></div></div>
        </div>
      </div>
    </motion.div>
  );
}

function CounterStat({ value, label, suffix = '' }: { value: number; label: string; suffix?: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  const { isDark } = useTheme();
  useEffect(() => {
    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting) {
        let start = 0; const step = value / 40;
        const timer = setInterval(() => {
          start += step;
          if (start >= value) { setCount(value); clearInterval(timer); } else { setCount(Math.floor(start)); }
        }, 30);
        observer.disconnect();
      }
    }, { threshold: 0.5 });
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [value]);
  return (
    <div ref={ref} className="text-center">
      <div className={`text-3xl sm:text-4xl font-bold ${isDark ? 'bg-gradient-to-r from-white to-white/70' : 'bg-gradient-to-r from-vortex-700 to-vortex-500'} bg-clip-text text-transparent`}>
        {count}{suffix}
      </div>
      <div className="text-sm text-themed-muted mt-1">{label}</div>
    </div>
  );
}

export default function HeroSection() {
  const { isDark } = useTheme();
  const { t } = useLang();
  return (
    <section className="relative min-h-screen flex flex-col justify-center overflow-hidden hero-gradient">
      <ParticleField isDark={isDark} />
      <div className={`absolute top-1/4 left-1/4 w-96 h-96 ${isDark ? 'bg-vortex-500/10' : 'bg-vortex-500/5'} rounded-full blur-3xl animate-float`} />
      <div className={`absolute bottom-1/4 right-1/4 w-80 h-80 ${isDark ? 'bg-cyber-400/8' : 'bg-cyber-400/3'} rounded-full blur-3xl animate-float-delayed`} />

      <div className="relative z-10 max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-32 pb-20">
        {/* Badge */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.5 }} className="flex justify-center mb-8">
          <a href="https://github.com/iPmartNetwork/VortexUI/releases" target="_blank" rel="noopener noreferrer"
            className={`inline-flex items-center gap-2 px-4 py-2 rounded-full glass text-sm group ${isDark ? 'border-vortex-500/20' : 'border-vortex-200'}`}>
            <Sparkles className="w-4 h-4 text-vortex-400" />
            <span className="text-themed-muted">{t('hero.badge.version')}</span>
            <span className="text-vortex-500 font-medium">{t('hero.badge.refresh')}</span>
            <ArrowRight className="w-3.5 h-3.5 text-themed-faint group-hover:translate-x-1 transition-transform" />
          </a>
        </motion.div>

        {/* Title */}
        <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6, delay: 0.1 }} className="text-center mb-6">
          <h1 className="text-5xl sm:text-6xl lg:text-8xl font-black tracking-tight leading-[1.1]">
            <span className={`${isDark ? 'bg-gradient-to-b from-white via-white to-white/40' : 'bg-gradient-to-b from-gray-900 via-gray-800 to-gray-500'} bg-clip-text text-transparent`}>
              {t('hero.title.line1')}
            </span>
            <br />
            <span className="bg-gradient-to-r from-vortex-400 via-cyber-400 to-vortex-400 bg-clip-text text-transparent animate-gradient bg-[length:200%_auto]">
              {t('hero.title.line2')}
            </span>
          </h1>
        </motion.div>

        {/* Subtitle */}
        <motion.p initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6, delay: 0.2 }}
          className="text-center text-lg sm:text-xl text-themed-muted max-w-3xl mx-auto mb-10 leading-relaxed">
          {t('hero.subtitle')}{' '}
          <span className="text-themed-secondary font-medium">Xray-core</span> {t('hero.subtitle.and')}{' '}
          <span className="text-themed-secondary font-medium">sing-box</span>.{' '}
          {t('hero.subtitle.rest')}{' '}
          <span className="text-vortex-500">{t('hero.subtitle.glass')}</span>{' '}
          {t('hero.subtitle.interface')}
        </motion.p>

        {/* CTA */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6, delay: 0.3 }}
          className="flex flex-col sm:flex-row items-center justify-center gap-4 mb-14">
          <a href="#install" className="relative group w-full sm:w-auto px-8 py-4 rounded-2xl text-base font-semibold text-white overflow-hidden flex items-center justify-center gap-2">
            <div className="absolute inset-0 bg-gradient-to-r from-vortex-600 to-vortex-500 group-hover:from-vortex-500 group-hover:to-cyber-500 transition-all duration-500" />
            <Terminal className="w-5 h-5 relative z-10" /><span className="relative z-10">{t('hero.cta.install')}</span>
            <ArrowRight className="w-4 h-4 relative z-10 group-hover:translate-x-1 rtl:group-hover:-translate-x-1 transition-transform" />
          </a>
          <a href="https://ipmartnetwork.github.io/VortexUI/" target="_blank" rel="noopener noreferrer"
            className={`w-full sm:w-auto px-8 py-4 rounded-2xl text-base font-semibold glass flex items-center justify-center gap-2 border transition-all ${
              isDark ? 'text-white/80 border-white/10 hover:border-white/20 hover:bg-white/10' : 'text-gray-700 border-gray-200 hover:border-vortex-300 hover:bg-vortex-50'
            }`}>
            {t('hero.cta.docs')}
          </a>
          <a href="https://github.com/iPmartNetwork/VortexUI" target="_blank" rel="noopener noreferrer"
            className={`w-full sm:w-auto px-8 py-4 rounded-2xl text-base font-semibold flex items-center justify-center gap-2 ${
              isDark ? 'text-white/60 hover:text-white' : 'text-gray-500 hover:text-gray-900'
            }`}>
            <GithubIcon className="w-5 h-5" /> {t('hero.cta.github')}
          </a>
        </motion.div>

        <TerminalDemo />

        {/* Stats */}
        <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 0.6, delay: 0.8 }}
          className="mt-20 grid grid-cols-2 sm:grid-cols-4 gap-8 max-w-3xl mx-auto">
          <CounterStat value={14} suffix="+" label={t('hero.stat.protocols')} />
          <CounterStat value={8} label={t('hero.stat.languages')} />
          <CounterStat value={6} label={t('hero.stat.formats')} />
          <CounterStat value={4} label={t('hero.stat.gateways')} />
        </motion.div>

        {/* Floating Icons */}
        <div className="hidden lg:block absolute top-40 left-10 animate-float">
          <div className="w-14 h-14 rounded-2xl glass flex items-center justify-center"><Shield className="w-7 h-7 text-vortex-400/60" /></div>
        </div>
        <div className="hidden lg:block absolute bottom-40 right-10 animate-float-delayed">
          <div className="w-14 h-14 rounded-2xl glass flex items-center justify-center"><Globe className="w-7 h-7 text-cyber-400/60" /></div>
        </div>
        <div className="hidden lg:block absolute top-60 right-20 animate-float">
          <div className="w-12 h-12 rounded-xl glass flex items-center justify-center"><Zap className="w-6 h-6 text-aurora-orange/60" /></div>
        </div>
      </div>

      <div className={`absolute bottom-0 left-0 right-0 h-40 bg-gradient-to-t ${isDark ? 'from-[#030014]' : 'from-[#f8f9fc]'} to-transparent z-10`} />
    </section>
  );
}
