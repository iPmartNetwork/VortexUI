import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Globe, Plus, Trash2 } from "lucide-react";
import { api } from "@/api/client";
import { Button } from "@/components/ui";
import { GlassCard } from "@/components/veltrix";
import { useToast } from "@/components/toast";
import { useTitle } from "@/lib/useTitle";

interface GeoExit { id: number; service: string; country: string; }

const COUNTRY_NAMES: Record<string,string> = { at:"Austria",be:"Belgium",bg:"Bulgaria",ca:"Canada",ch:"Switzerland",cz:"Czechia",de:"Germany",dk:"Denmark",ee:"Estonia",es:"Spain",fi:"Finland",fr:"France",gb:"UK",hu:"Hungary",ie:"Ireland",is:"Iceland",it:"Italy",lt:"Lithuania",lu:"Luxembourg",lv:"Latvia",nl:"Netherlands",no:"Norway",pl:"Poland",pt:"Portugal",ro:"Romania",rs:"Serbia",se:"Sweden",si:"Slovenia",sk:"Slovakia",tr:"Turkey",ua:"Ukraine",us:"USA" };

export function GeoExits() {
  useTitle("Geo Exits");
  const qc = useQueryClient();
  const toast = useToast();
  const [svc, setSvc] = useState("tor");
  const [country, setCountry] = useState("");

  const { data } = useQuery({ queryKey: ["geo-exits"], queryFn: () => api<{ exits: GeoExit[] }>("/api/v2/geo-exits") });
  const { data: regions } = useQuery({ queryKey: ["geo-regions", svc], queryFn: () => api<{ countries: string[] }>(`/api/v2/geo-exits/regions/${svc}`) });

  const addMut = useMutation({ mutationFn: (body: { service: string; country: string }) => api("/api/v2/geo-exits", { method: "POST", body }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["geo-exits"] }); toast.success("Geo exit added"); } });
  const delMut = useMutation({ mutationFn: (id: number) => api(`/api/v2/geo-exits/${id}`, { method: "DELETE" }), onSuccess: () => { qc.invalidateQueries({ queryKey: ["geo-exits"] }); toast.success("Removed"); } });

  return (
    <div className="space-y-6 p-6">
      <h1 className="text-2xl font-bold flex items-center gap-2"><Globe className="w-6 h-6" />Geo Exits (Tor / Psiphon)</h1>
      <p className="text-sm text-fg-muted">Route traffic through specific countries using Tor or Psiphon exit nodes. Useful when direct routes are blocked.</p>

      <GlassCard className="p-4 space-y-3">
        <h3 className="font-semibold">Add Exit</h3>
        <div className="flex gap-3 items-end flex-wrap">
          <div>
            <label className="text-xs font-medium">Service</label>
            <select className="field input-surface w-full mt-1" value={svc} onChange={(e) => { setSvc(e.target.value); setCountry(""); }}>
              <option value="tor">Tor</option>
              <option value="psiphon">Psiphon</option>
            </select>
          </div>
          <div>
            <label className="text-xs font-medium">Country</label>
            <select className="field input-surface w-full mt-1" value={country} onChange={(e) => setCountry(e.target.value)}>
              <option value="">Select...</option>
              {(regions?.countries ?? []).map((cc) => <option key={cc} value={cc}>{COUNTRY_NAMES[cc] || cc.toUpperCase()} ({cc})</option>)}
            </select>
          </div>
          <Button disabled={!country} onClick={() => addMut.mutate({ service: svc, country })}><Plus size={14} /> Add</Button>
        </div>
      </GlassCard>

      <GlassCard className="p-4">
        <h3 className="font-semibold mb-3">Active Exits</h3>
        {(data?.exits ?? []).length === 0 ? <p className="text-sm text-fg-muted">No geo exits configured.</p> : (
          <div className="space-y-2">
            {(data?.exits ?? []).map((e) => (
              <div key={e.id} className="flex items-center justify-between border border-border rounded-lg px-3 py-2">
                <span className="text-sm font-medium">{e.service.toUpperCase()} → {COUNTRY_NAMES[e.country] || e.country} ({e.country})</span>
                <Button variant="ghost" size="sm" className="text-red-500" onClick={() => delMut.mutate(e.id)}><Trash2 size={14} /></Button>
              </div>
            ))}
          </div>
        )}
      </GlassCard>
    </div>
  );
}
