import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "./client";

export interface ResellerPaymentConfig {
  card_number: string;
  card_holder: string;
  card_bank: string;
  crypto_addresses: Record<string, string>;
  zarinpal_merchant_id: string;
  manual_instructions: string;
  enabled_methods: string[];
}

export interface PendingOrder {
  id: string;
  user_id: string;
  admin_id: string;
  plan_id: string;
  username: string;
  status: string;
  gateway: string;
  gateway_id: string;
  amount: number;
  currency: string;
  proof_image?: string;
  created_at: string;
  paid_at: string | null;
}

export function useResellerPaymentConfig() {
  return useQuery({
    queryKey: ["reseller-payment-config"],
    queryFn: () => api<{ config: ResellerPaymentConfig }>("/api/account/payment-config"),
  });
}

export function useSaveResellerPaymentConfig() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (config: Omit<ResellerPaymentConfig, "enabled_methods"> & { enabled_methods: string[] }) =>
      api("/api/account/payment-config", { method: "PUT", body: config }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ["reseller-payment-config"] }),
  });
}

export function usePendingOrders() {
  return useQuery({
    queryKey: ["pending-orders"],
    queryFn: () => api<{ orders: PendingOrder[] }>("/api/orders/pending"),
    refetchInterval: 15000,
  });
}

export function useReviewOrder() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (args: { id: string; action: "approve" | "reject"; note?: string }) =>
      api<{ status: string; order_id: string }>(`/api/orders/${args.id}/review`, {
        method: "POST",
        body: { action: args.action, note: args.note ?? "" },
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["pending-orders"] });
      qc.invalidateQueries({ queryKey: ["orders"] });
    },
  });
}
