import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";

export interface WalletPackage {
  id: string;
  name: string;
  description?: string;
  traffic_bytes: number;
  user_credits: number;
  price_amount: number;
  currency: string;
  methods: string[];
  enabled: boolean;
  sort_order: number;
}

export interface BillingSettings {
  card_number: string;
  card_holder: string;
  card_bank: string;
  crypto_addresses: Record<string, string>;
  manual_instructions: string;
}

export interface WalletDeposit {
  id: string;
  admin_id: string;
  admin_username?: string;
  package_id?: string;
  package_name?: string;
  method: string;
  status: string;
  amount: number;
  currency: string;
  traffic_bytes: number;
  user_credits: number;
  tx_id?: string;
  proof_image?: string;
  reseller_note?: string;
  admin_note?: string;
  created_at: string;
  reviewed_at?: string;
  paid_at?: string;
}

export function useWalletPackages(enabledOnly = false) {
  const path = enabledOnly ? "/api/account/wallet/packages" : "/api/billing/wallet-packages";
  return useQuery({
    queryKey: ["wallet-packages", enabledOnly],
    queryFn: () => api<{ packages: WalletPackage[] }>(path),
  });
}

export function usePaymentInfo() {
  return useQuery({
    queryKey: ["wallet-payment-info"],
    queryFn: () =>
      api<{ settings: BillingSettings; zarinpal_enabled: boolean; crypto_enabled: boolean }>(
        "/api/account/wallet/payment-info",
      ),
  });
}

export function useAccountDeposits() {
  return useQuery({
    queryKey: ["account-wallet-deposits"],
    queryFn: () => api<{ deposits: WalletDeposit[] }>("/api/account/wallet/deposits"),
  });
}

export function useBillingDeposits(status?: string) {
  const q = status ? `?status=${encodeURIComponent(status)}` : "";
  return useQuery({
    queryKey: ["billing-deposits", status],
    queryFn: () => api<{ deposits: WalletDeposit[] }>(`/api/billing/deposits${q}`),
  });
}

export function useBillingSettings() {
  return useQuery({
    queryKey: ["billing-settings"],
    queryFn: () => api<{ settings: BillingSettings }>("/api/billing/settings"),
  });
}

export function useInitWalletDeposit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: {
      package_id: string;
      method: string;
      tx_id?: string;
      proof_image?: string;
      reseller_note?: string;
    }) => api<{ deposit: WalletDeposit; redirect_url?: string }>("/api/account/wallet/deposits", { method: "POST", body }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["account-wallet-deposits"] });
      qc.invalidateQueries({ queryKey: ["account-wallet"] });
    },
  });
}

export function useReviewDeposit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; action: "approve" | "reject"; note?: string }) =>
      api(`/api/billing/deposits/${args.id}/review`, {
        method: "POST",
        body: { action: args.action, note: args.note },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["billing-deposits"] });
      qc.invalidateQueries({ queryKey: ["reseller-quota-usage"] });
    },
  });
}

export function useSaveBillingSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (settings: BillingSettings) =>
      api("/api/billing/settings", { method: "PUT", body: settings }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["billing-settings"] }),
  });
}

export function useCreateWalletPackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (body: Partial<WalletPackage>) =>
      api("/api/billing/wallet-packages", { method: "POST", body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["wallet-packages"] }),
  });
}

export function useUpdateWalletPackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (pkg: WalletPackage) =>
      api(`/api/billing/wallet-packages/${pkg.id}`, { method: "PUT", body: pkg }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["wallet-packages"] }),
  });
}

export function useDeleteWalletPackage() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => api(`/api/billing/wallet-packages/${id}`, { method: "DELETE" }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["wallet-packages"] }),
  });
}

export function formatPrice(amount: number, currency: string) {
  const cur = currency.toUpperCase();
  if (cur === "IRR") return `${amount.toLocaleString()} Toman`;
  if (cur === "USD") return `$${(amount / 100).toFixed(2)}`;
  return `${amount} ${cur}`;
}
