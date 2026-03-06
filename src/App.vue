<script setup lang="ts">
import { ref, computed } from 'vue';
import { invoke } from "@tauri-apps/api/core";

// Define TypeScript interfaces for strict production typing
interface Device {
  ip: string;
  mac: string;
  status: string;
}

interface NetworkReport {
  devices: Device[];
  scanMethod: string;
  subnet: string;
  timestamp: string;
  performance: string;
}

// State Management
const targetCidr = ref("192.168.1.0/24");
const community = ref("");
const report = ref<NetworkReport | null>(null);
const isScanning = ref(false);
const searchQuery = ref("");

// Computed: Filter results in real-time
const filteredDevices = computed(() => {
  if (!report.value) return [];
  return report.value.devices.filter(d => 
    d.ip.includes(searchQuery.value) || d.mac.toLowerCase().includes(searchQuery.value.toLowerCase())
  );
});

async function startAudit() {
  if (isScanning.value) return;
  isScanning.value = true;
  report.value = null; // Clear old data

  try {
    // Calling the Rust bridge which executes the Go Sidecar
    const result = await invoke<string>("execute_enterprise_audit", { 
      args: { 
        target: targetCidr.value, 
        community: community.value || null 
      } 
    });
    report.value = JSON.parse(result);
  } catch (error) {
    alert("Audit Failed: Check CIDR format or Admin privileges.");
    console.error(error);
  } finally {
    isScanning.value = false;
  }
}

const exportCSV = () => {
  if (!report.value) return;
  const content = "IP,MAC,Status\n" + report.value.devices.map(d => `${d.ip},${d.mac},${d.status}`).join("\n");
  const blob = new Blob([content], { type: 'text/csv' });
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `network-audit-${Date.now()}.csv`;
  a.click();
};
</script>

<template>
  <main class="min-h-screen bg-[#0a0a0c] text-slate-200 font-sans selection:bg-blue-500/30">
    <nav class="border-b border-white/5 bg-black/40 backdrop-blur-md sticky top-0 z-50">
      <div class="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
            <span class="font-black text-white text-xs">LM</span>
          </div>
          <h1 class="font-bold tracking-tight text-lg uppercase italic">LAN <span class="text-blue-500">Maintainer</span></h1>
        </div>
        <div v-if="report" class="flex gap-4">
          <button @click="exportCSV" class="text-xs bg-white/5 hover:bg-white/10 px-4 py-2 rounded-full border border-white/10 transition-all">
            Export CSV
          </button>
        </div>
      </div>
    </nav>

    <div class="max-w-7xl mx-auto px-6 py-8">
      <div class="grid grid-cols-12 gap-8">
        
        <aside class="col-span-12 lg:col-span-4 space-y-6">
          <section class="bg-[#111114] p-6 rounded-2xl border border-white/5 shadow-2xl">
            <h2 class="text-xs font-black uppercase tracking-widest text-slate-500 mb-6">Audit Configuration</h2>
            
            <div class="space-y-4">
              <div>
                <label class="block text-xs font-medium text-slate-400 mb-2 ml-1">Subnet Target (CIDR)</label>
                <input v-model="targetCidr" 
                  class="w-full bg-black/50 border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all font-mono"
                  placeholder="192.168.1.0/24" />
              </div>

              <div>
                <label class="block text-xs font-medium text-slate-400 mb-2 ml-1">SNMP Community (Managed Switch Only)</label>
                <input v-model="community" type="password"
                  class="w-full bg-black/50 border border-white/10 rounded-xl px-4 py-3 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all font-mono"
                  placeholder="••••••••" />
              </div>

              <button @click="startAudit" :disabled="isScanning"
                class="w-full py-4 rounded-xl font-bold text-sm tracking-widest uppercase transition-all flex items-center justify-center gap-3 mt-4"
                :class="isScanning ? 'bg-slate-800 text-slate-500 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/20 active:scale-[0.98]'">
                <span v-if="isScanning" class="w-4 h-4 border-2 border-white/20 border-t-white rounded-full animate-spin"></span>
                {{ isScanning ? 'Executing Probe...' : 'Execute Audit' }}
              </button>
            </div>
          </section>

          <div v-if="report" class="bg-blue-600/5 border border-blue-500/20 p-5 rounded-2xl">
            <h3 class="text-[10px] font-black uppercase text-blue-400 mb-1">Methodology</h3>
            <p class="text-sm font-semibold">{{ report.scanMethod }}</p>
            <p class="text-[10px] text-slate-500 mt-2 italic">Last scan: {{ report.timestamp }}</p>
          </div>
        </aside>

        <div class="col-span-12 lg:col-span-8 space-y-6">
          <div v-if="!report && !isScanning" class="h-64 flex flex-col items-center justify-center border-2 border-dashed border-white/5 rounded-3xl opacity-50">
             <p class="text-sm font-medium">System Idle. Awaiting CIDR Input.</p>
          </div>

          <div v-if="report" class="space-y-4">
            <div class="flex items-center justify-between gap-4">
              <input v-model="searchQuery" 
                class="bg-[#111114] border border-white/5 rounded-full px-6 py-2 text-sm w-full max-w-md focus:border-blue-500 outline-none transition-all"
                placeholder="Filter by IP or MAC..." />
              <div class="text-xs font-mono text-slate-500 whitespace-nowrap">
                Found {{ filteredDevices.length }} Nodes
              </div>
            </div>

            <div class="bg-[#111114] border border-white/5 rounded-2xl overflow-hidden shadow-2xl">
              <table class="w-full text-left">
                <thead class="bg-white/5">
                  <tr class="text-[10px] uppercase tracking-widest text-slate-500">
                    <th class="px-6 py-4 font-black">Endpoint Address</th>
                    <th class="px-6 py-4 font-black">Hardware Hash (MAC)</th>
                    <th class="px-6 py-4 font-black">Link Status</th>
                  </tr>
                </thead>
                <tbody class="divide-y divide-white/5">
                  <tr v-for="device in filteredDevices" :key="device.ip" class="hover:bg-white/[0.02] transition-colors group">
                    <td class="px-6 py-4 text-sm font-mono text-blue-400 group-hover:text-blue-300">{{ device.ip }}</td>
                    <td class="px-6 py-4 text-sm font-mono text-slate-400">{{ device.mac }}</td>
                    <td class="px-6 py-4">
                      <span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold uppercase bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
                        <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                        Active
                      </span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

      </div>
    </div>
  </main>
</template>