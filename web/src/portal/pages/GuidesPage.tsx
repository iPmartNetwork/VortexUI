import { useQuery } from "@tanstack/react-query";
import { BookOpen } from "lucide-react";
import { api } from "@/api/client";
import { GlassCard } from "@/components/veltrix";
import { useI18n } from "@/i18n/i18n";

interface ConnectionGuide {
  id: string;
  app_name: string;
  platform: string;
  icon_url: string;
  content: string;
}

export function GuidesPage() {
  const { t } = useI18n();

  const { data: guides } = useQuery({
    queryKey: ["portal-guides"],
    queryFn: () =>
      api.get("/api/v2/portal/guides").then((r) => r.data as ConnectionGuide[]),
  });

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-bold flex items-center gap-2">
        <BookOpen className="w-5 h-5" />
        {t("portal.guides", "Connection Guides")}
      </h2>
      <p className="text-muted-foreground">
        {t("portal.guidesDesc", "Step-by-step instructions for your client app.")}
      </p>

      {guides && guides.length > 0 ? (
        <div className="grid gap-4 sm:grid-cols-2">
          {guides.map((guide) => (
            <GlassCard key={guide.id} className="p-4">
              <div className="flex items-center gap-3 mb-3">
                {guide.icon_url && (
                  <img src={guide.icon_url} alt="" className="w-8 h-8 rounded" />
                )}
                <div>
                  <h3 className="font-medium">{guide.app_name}</h3>
                  <span className="text-xs text-muted-foreground">
                    {guide.platform}
                  </span>
                </div>
              </div>
              <div
                className="prose prose-sm dark:prose-invert"
                dangerouslySetInnerHTML={{ __html: guide.content }}
              />
            </GlassCard>
          ))}
        </div>
      ) : (
        <p className="text-muted-foreground text-sm">
          {t("portal.noGuides", "No guides available yet.")}
        </p>
      )}
    </div>
  );
}
