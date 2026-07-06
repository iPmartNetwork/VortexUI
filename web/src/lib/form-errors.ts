import type { TKey } from "@/i18n/dict";

export interface FormErrorLike {
  status?: number;
  message?: string;
}

export type TranslateFn = (key: TKey) => string;

export function getApiErrorMessage(error: unknown, fallback: string, t?: TranslateFn): string {
  if (!error || typeof error !== "object") return fallback;

  const maybe = error as FormErrorLike;
  if (typeof maybe.message === "string" && maybe.message.trim()) {
    const msg = maybe.message.trim();
    // Skip generic client-side placeholders when a translator is available.
    if (t && (msg === "unauthorized" || msg.startsWith("request failed ("))) {
      // fall through to status-based message
    } else {
      return msg;
    }
  }

  if (!t) {
    switch (maybe.status) {
      case 400:
        return "The submitted data is invalid.";
      case 401:
        return "Your session has expired. Please sign in again.";
      case 403:
        return "You do not have permission to perform this action.";
      case 404:
        return "The requested resource could not be found.";
      case 409:
        return "This change conflicts with an existing item.";
      case 429:
        return "Too many requests. Please wait a moment and try again.";
      default:
        return fallback;
    }
  }

  switch (maybe.status) {
    case 400:
      return t("errors.invalidData");
    case 401:
      return t("errors.sessionExpired");
    case 403:
      return t("errors.forbidden");
    case 404:
      return t("errors.notFound");
    case 409:
      return t("errors.conflict");
    case 429:
      return t("errors.tooManyRequests");
    default:
      return fallback;
  }
}
