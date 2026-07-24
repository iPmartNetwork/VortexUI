import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { BookOpen, BarChart3, ExternalLink } from "lucide-react";
import { api } from "@/api/client";
import { Button, Badge } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useTitle } from "@/lib/useTitle";
import { useI18n } from "@/i18n/i18n";

interface RateLimitInfo {
  endpoint: string;
  limit: number;
  remaining: number;
  window_sec: number;
  reset_at: number;
}

export function ApiDocs() {
  const { t } = useI18n();
  useTitle(t("apiDocs.title", "API Documentation"));
  const [activeTab, setActiveTab] = useState<"docs" | "rate-limits">("docs");

  const { data: rateLimits } = useQuery({
    queryKey: ["rate-limits"],
    queryFn: () => api.get("/api/v2/rate-limits").then((r) => r.data as RateLimitInfo[]),
    enabled: activeTab === "rate-limits",
    refetchInterval: 5000,
  });

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <BookOpen className="w-6 h-6" />
        {t("apiDocs.title", "API Documentation")}
      </h1>

      {/* Tabs */}
      <div className="flex gap-2 border-b pb-2">
        <Button
          variant={activeTab === "docs" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("docs")}
        >
          <BookOpen className="w-4 h-4 mr-1" />
          {t("apiDocs.documentation", "Documentation")}
        </Button>
        <Button
          variant={activeTab === "rate-limits" ? "default" : "ghost"}
          size="sm"
          onClick={() => setActiveTab("rate-limits")}
        >
          <BarChart3 className="w-4 h-4 mr-1" />
          {t("apiDocs.rateLimits", "Rate Limits")}
        </Button>
      </div>

      {activeTab === "docs" && (
        <GlassCard className="p-6 space-y-4">
          <h2 className="text-lg font-semibold">{t("apiDocs.swaggerUI", "Swagger UI")}</h2>
          <p className="text-muted-foreground">
            {t("apiDocs.swaggerDesc", "Interactive API documentation with Try-it-out functionality.")}
          </p>
          <div className="flex gap-3">
            <a href="/api/v2/docs/" target="_blank" rel="noopener noreferrer">
              <Button>
                <ExternalLink className="w-4 h-4 mr-1" />
                {t("apiDocs.openSwagger", "Open Swagger UI")}
              </Button>
            </a>
            <a href="/api/v2/docs/openapi.yaml" target="_blank" rel="noopener noreferrer">
              <Button variant="outline">
                {t("apiDocs.downloadSpec", "Download OpenAPI Spec")}
              </Button>
            </a>
          </div>

          <hr className="my-4" />

          <h2 className="text-lg font-semibold">{t("apiDocs.sdks", "Client SDKs")}</h2>
          <p className="text-muted-foreground">
            {t("apiDocs.sdkDesc", "Auto-generated client libraries from the OpenAPI specification.")}
          </p>
          <div className="grid grid-cols-3 gap-3">
            {["Python", "JavaScript", "Go"].map((lang) => (
              <div key={lang} className="border rounded p-3 text-center">
                <span className="font-medium">{lang}</span>
                <p className="text-xs text-muted-foreground mt-1">
                  {t("apiDocs.generatedFrom", "Generated from OpenAPI")}
                </p>
              </div>
            ))}
          </div>

          <hr className="my-4" />

          <h2 className="text-lg font-semibold">{t("apiDocs.versioning", "API Versioning")}</h2>
          <p className="text-muted-foreground">
            {t("apiDocs.versioningDesc", "The API supports versioned endpoints. Current version: v2. Legacy v1 endpoints remain functional.")}
          </p>
          <div className="flex gap-2">
            <Badge>v2 (current)</Badge>
            <Badge variant="outline">v1 (legacy)</Badge>
          </div>
        </GlassCard>
      )}

      {activeTab === "rate-limits" && (
        <GlassCard className="p-4 space-y-4">
          <h2 className="text-lg font-semibold flex items-center gap-2">
            <BarChart3 className="w-5 h-5" />
            {t("apiDocs.rateLimitDashboard", "Rate Limit Dashboard")}
          </h2>
          {rateLimits && rateLimits.length > 0 ? (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3">Endpoint</th>
                    <th className="text-left py-2 px-3">Limit</th>
                    <th className="text-left py-2 px-3">Remaining</th>
                    <th className="text-left py-2 px-3">Window</th>
                    <th className="text-left py-2 px-3">Usage</th>
                  </tr>
                </thead>
                <tbody>
                  {rateLimits.map((rl) => {
                    const usage = ((rl.limit - rl.remaining) / rl.limit) * 100;
                    return (
                      <tr key={rl.endpoint} className="border-b hover:bg-muted/50">
                        <td className="py-2 px-3 font-mono text-xs">{rl.endpoint}</td>
                        <td className="py-2 px-3">{rl.limit}</td>
                        <td className="py-2 px-3">
                          <Badge variant={rl.remaining < rl.limit * 0.2 ? "destructive" : "default"}>
                            {rl.remaining}
                          </Badge>
                        </td>
                        <td className="py-2 px-3">{rl.window_sec}s</td>
                        <td className="py-2 px-3">
                          <div className="w-24 h-2 bg-muted rounded-full overflow-hidden">
                            <div
                              className={`h-full rounded-full ${
                                usage > 80 ? "bg-destructive" : usage > 50 ? "bg-yellow-500" : "bg-green-500"
                              }`}
                              style={{ width: `${Math.min(usage, 100)}%` }}
                            />
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          ) : (
            <p className="text-muted-foreground text-sm">
              {t("apiDocs.noRateLimits", "No rate limit data available.")}
            </p>
          )}
        </GlassCard>
      )}
    </div>
  );
}
