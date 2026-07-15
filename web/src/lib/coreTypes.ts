export type CoreType = "xray" | "singbox";

export function normalizedEnabledCores(node: {
  core: CoreType;
  enabled_cores?: CoreType[];
}): CoreType[] {
  if (node.enabled_cores && node.enabled_cores.length > 0) return node.enabled_cores;
  return [node.core || "xray"];
}

export function isMultiCore(node: { core: CoreType; enabled_cores?: CoreType[] }): boolean {
  return normalizedEnabledCores(node).length > 1;
}

export function resolveInboundCore(
  node: { core: CoreType },
  override?: CoreType | "",
): CoreType {
  if (override) return override;
  return node.core || "xray";
}

export function inboundEffectiveCore(
  node: { core: CoreType },
  inbound: { core?: CoreType | "" },
): CoreType {
  return resolveInboundCore(node, inbound.core);
}
