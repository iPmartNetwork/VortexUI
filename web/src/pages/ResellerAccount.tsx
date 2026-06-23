import { useEffect, useState } from "react";
import { Wallet, Users, Palette, Webhook } from "lucide-react";
import { useRoles } from "@/api/admin-hooks";
import {
  useAccountBranding,
  useAccountWallet,
  useAccountWebhook,
  useCreateSubAdmin,
  useSaveBranding,
  useSaveWebhook,
  useSubAdmins,
} from "@/api/reseller-hooks";
import { useAuth } from "@/auth/auth";
import { Badge, Button, Card, Input, PageHeader } from "@/components/ui";
import { useToast } from "@/components/toast";
import { useI18n } from "@/i18n/i18n";
import { formatBytes } from "@/lib/utils";

export function ResellerAccount() {
  const { sudo, session } = useAuth();
  const { t } = useI18n();
  const toast = useToast();
  const wallet = useAccountWallet();
  const subAdmins = useSubAdmins();
  const brandingQ = useAccountBranding();
  const webhookQ = useAccountWebhook();
  const roles = useRoles();
  const saveBranding = useSaveBranding();
  const saveWebhook = useSaveWebhook();
  const createSub = useCreateSubAdmin();

  const [branding, setBranding] = useState(brandingQ.data?.branding);
  const [webhookUrl, setWebhookUrl] = useState("");
  const [webhookSecret, setWebhookSecret] = useState("");
  const [webhookEnabled, setWebhookEnabled] = useState(false);
  const [subForm, setSubForm] = useState({
    username: "",
    password: "",
    role_id: "",
    user_quota: 0,
    traffic_quota: 0,
    traffic_quota_mode: "allocated",
  });

  useEffect(() => {
    if (brandingQ.data?.branding) setBranding(brandingQ.data.branding);
  }, [brandingQ.data]);

  useEffect(() => {
    if (webhookQ.data) {
      setWebhookUrl(webhookQ.data.url);
      setWebhookEnabled(webhookQ.data.enabled);
    }
  }, [webhookQ.data]);

  if (sudo) {
    return (
      <div className="space-y-4">
        <PageHeader title={t("reseller.account.title")} subtitle={t("reseller.account.sudoHint")} />
        <Card className="p-6 text-sm text-muted-foreground">{t("reseller.account.sudoCard")}</Card>
      </div>
    );
  }

  const w = wallet.data?.wallet;
  const ledger = wallet.data?.ledger ?? [];

  async function onSaveBranding() {
    if (!branding) return;
    try {
      await saveBranding.mutateAsync(branding);
      toast.success(t("reseller.account.brandingSaved"));
    } catch {
      toast.error(t("reseller.account.brandingFailed"));
    }
  }

  async function onSaveWebhook() {
    try {
      await saveWebhook.mutateAsync({ url: webhookUrl, secret: webhookSecret, enabled: webhookEnabled });
      setWebhookSecret("");
      toast.success(t("reseller.account.webhookSaved"));
    } catch {
      toast.error(t("reseller.account.webhookFailed"));
    }
  }

  async function onCreateSub() {
    try {
      await createSub.mutateAsync(subForm);
      toast.success(t("reseller.account.subCreated"));
      setSubForm({ username: "", password: "", role_id: "", user_quota: 0, traffic_quota: 0, traffic_quota_mode: "allocated" });
    } catch (e) {
      toast.error(e instanceof Error ? e.message : t("reseller.account.createFailed"));
    }
  }

  const allowSubResellers = !!session?.admin.allow_sub_resellers;

  return (
    <div className="space-y-8 animate-fade-in">
      <PageHeader title={t("reseller.account.title")} subtitle={t("reseller.account.subtitle")} />

      <section className="space-y-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold"><Wallet size={18} /> {t("reseller.account.wallet")}</h2>
        <div className="grid gap-4 sm:grid-cols-2">
          <Card className="p-5">
            <div className="text-xs text-muted-foreground">{t("reseller.account.trafficCredits")}</div>
            <div className="text-2xl font-bold">{w ? formatBytes(w.traffic_bytes, false) : "—"}</div>
          </Card>
          <Card className="p-5">
            <div className="text-xs text-muted-foreground">{t("reseller.account.userCredits")}</div>
            <div className="text-2xl font-bold">{w?.user_credits ?? "—"}</div>
          </Card>
        </div>
        {ledger.length > 0 && (
          <Card className="p-0 overflow-x-auto">
            <table className="w-full text-sm">
              <thead className="border-b text-left text-muted-foreground">
                <tr>
                  <th className="px-4 py-2">{t("reseller.account.when")}</th>
                  <th className="px-4 py-2">{t("reseller.account.traffic")}</th>
                  <th className="px-4 py-2">{t("reseller.account.users")}</th>
                  <th className="px-4 py-2">{t("reseller.account.reason")}</th>
                </tr>
              </thead>
              <tbody>
                {ledger.map((e) => (
                  <tr key={e.id} className="border-b last:border-0">
                    <td className="px-4 py-2 text-muted-foreground">{new Date(e.created_at).toLocaleString()}</td>
                    <td className="px-4 py-2">{e.delta_traffic ? formatBytes(e.delta_traffic, false) : "—"}</td>
                    <td className="px-4 py-2">{e.delta_users || "—"}</td>
                    <td className="px-4 py-2">{e.reason}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </Card>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold"><Users size={18} /> {t("reseller.account.subResellers")}</h2>
        <Card className="p-0">
          <table className="w-full text-sm">
            <thead className="border-b text-left text-muted-foreground">
              <tr>
                <th className="px-4 py-2">{t("reseller.account.colUsername")}</th>
                <th className="px-4 py-2">{t("reseller.account.colUserQuota")}</th>
                <th className="px-4 py-2">{t("reseller.account.colTrafficQuota")}</th>
              </tr>
            </thead>
            <tbody>
              {(subAdmins.data?.admins ?? []).map((a) => (
                <tr key={a.id} className="border-b last:border-0">
                  <td className="px-4 py-2 font-medium">{a.username}</td>
                  <td className="px-4 py-2">{a.user_quota || "∞"}</td>
                  <td className="px-4 py-2">{a.traffic_quota ? formatBytes(a.traffic_quota, false) : "∞"}</td>
                </tr>
              ))}
              {(subAdmins.data?.admins ?? []).length === 0 && (
                <tr><td colSpan={3} className="px-4 py-4 text-muted-foreground">{t("reseller.account.noSubResellers")}</td></tr>
              )}
            </tbody>
          </table>
        </Card>
        {allowSubResellers && (
        <Card className="space-y-3 p-5">
          <div className="text-sm font-medium">{t("reseller.account.createSubReseller")}</div>
          <div className="grid gap-3 sm:grid-cols-2">
            <Input placeholder={t("reseller.account.ph.username")} value={subForm.username} onChange={(e) => setSubForm({ ...subForm, username: e.target.value })} />
            <Input type="password" placeholder={t("reseller.account.ph.password")} value={subForm.password} onChange={(e) => setSubForm({ ...subForm, password: e.target.value })} />
            <select className="input" value={subForm.role_id} onChange={(e) => setSubForm({ ...subForm, role_id: e.target.value })}>
              <option value="">{t("reseller.account.selectRole")}</option>
              {(roles.data?.roles ?? []).map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
            </select>
            <Input type="number" placeholder={t("reseller.account.ph.userQuota")} value={subForm.user_quota} onChange={(e) => setSubForm({ ...subForm, user_quota: Number(e.target.value) })} />
            <Input type="number" placeholder={t("reseller.account.ph.trafficQuota")} value={subForm.traffic_quota} onChange={(e) => setSubForm({ ...subForm, traffic_quota: Number(e.target.value) })} />
          </div>
          <Button onClick={onCreateSub} disabled={createSub.isPending}>{t("common.create")}</Button>
        </Card>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold"><Palette size={18} /> {t("reseller.account.branding")}</h2>
        {branding && (
          <Card className="grid gap-3 p-5 sm:grid-cols-2">
            <Input placeholder={t("reseller.account.ph.panelTitle")} value={branding.panel_title} onChange={(e) => setBranding({ ...branding, panel_title: e.target.value })} />
            <div className="space-y-2">
              <label className="block text-xs text-muted-foreground">{t("reseller.account.ph.logoUrl")}</label>
              <input
                type="file"
                accept="image/png,image/jpeg,image/webp,image/gif"
                className="block w-full text-sm"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (!file || file.size > 512_000) return;
                  const reader = new FileReader();
                  reader.onload = () => setBranding({ ...branding, logo_url: String(reader.result ?? "") });
                  reader.readAsDataURL(file);
                }}
              />
              {branding.logo_url && (
                <img src={branding.logo_url} alt="" className="h-12 max-w-[160px] rounded border object-contain" />
              )}
            </div>
            <div className="space-y-2">
              <label className="block text-xs text-muted-foreground">{t("reseller.account.ph.accentColor")}</label>
              <div className="flex items-center gap-2">
                <input
                  type="color"
                  value={/^#[0-9a-fA-F]{6}$/.test(branding.accent_color) ? branding.accent_color : "#6366f1"}
                  onChange={(e) => setBranding({ ...branding, accent_color: e.target.value })}
                  className="h-10 w-14 cursor-pointer rounded border bg-transparent"
                />
                <Input value={branding.accent_color} onChange={(e) => setBranding({ ...branding, accent_color: e.target.value })} />
              </div>
            </div>
            <Input placeholder={t("reseller.account.ph.portalSlug")} value={branding.portal_slug ?? ""} onChange={(e) => setBranding({ ...branding, portal_slug: e.target.value })} />
            <Input className="sm:col-span-2" placeholder={t("reseller.account.ph.footerText")} value={branding.footer_text} disabled readOnly />
            <Button onClick={onSaveBranding} disabled={saveBranding.isPending}>{t("reseller.account.saveBranding")}</Button>
          </Card>
        )}
      </section>

      <section className="space-y-3">
        <h2 className="flex items-center gap-2 text-lg font-semibold"><Webhook size={18} /> {t("reseller.account.webhook")}</h2>
        <Card className="space-y-3 p-5">
          <Input placeholder={t("reseller.account.ph.webhookUrl")} value={webhookUrl} onChange={(e) => setWebhookUrl(e.target.value)} />
          <Input
            type="password"
            placeholder={webhookQ.data?.has_secret ? t("reseller.account.newSecret") : t("reseller.account.signingSecret")}
            value={webhookSecret}
            onChange={(e) => setWebhookSecret(e.target.value)}
          />
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={webhookEnabled} onChange={(e) => setWebhookEnabled(e.target.checked)} />
            {t("common.enabled")}
            {webhookQ.data?.has_secret && <Badge>{t("reseller.account.secretSet")}</Badge>}
          </label>
          <p className="text-xs text-muted-foreground">{t("reseller.account.webhookHint")}</p>
          <Button onClick={onSaveWebhook} disabled={saveWebhook.isPending}>{t("reseller.account.saveWebhook")}</Button>
        </Card>
      </section>
    </div>
  );
}
