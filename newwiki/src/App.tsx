import { useTheme } from './contexts/ThemeContext';
import { useLang } from './contexts/LangContext';
import Navbar from './components/Navbar';
import HeroSection from './components/HeroSection';
import FeaturesSection from './components/FeaturesSection';
import DashboardPreview from './components/DashboardPreview';
import ArchitectureSection from './components/ArchitectureSection';
import ProtocolsSection from './components/ProtocolsSection';
import SecuritySection from './components/SecuritySection';
import ComparisonSection from './components/ComparisonSection';
import InstallSection from './components/InstallSection';
import DocumentationSection from './components/DocumentationSection';
import Footer from './components/Footer';

export default function App() {
  const { isDark } = useTheme();
  const { dir } = useLang();

  return (
    <div
      dir={dir}
      className={`noise-bg min-h-screen font-sans transition-colors duration-500 ${
        isDark
          ? 'bg-[#030014] text-white'
          : 'bg-[#f8f9fc] text-[#1a1a2e]'
      }`}
    >
      <Navbar />
      <main>
        <HeroSection />
        <FeaturesSection />
        <DashboardPreview />
        <ArchitectureSection />
        <ProtocolsSection />
        <SecuritySection />
        <ComparisonSection />
        <InstallSection />
        <DocumentationSection />
      </main>
      <Footer />
    </div>
  );
}
