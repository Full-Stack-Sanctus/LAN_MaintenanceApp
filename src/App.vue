<script setup lang="ts">
import { ref, computed } from 'vue';
import { invoke } from "@tauri-apps/api/core";

// Interfaces for strict production typing
interface Device {
  ip: string;
  mac: string;
  status: string;
  ttl?: number;
  os?: string;
  subnetMatch?: boolean;
}

interface NetworkReport {
  devices: Device[];
  scanMethod: string;
  subnet: string;
  timestamp: string;
  performance: string;
}


const targetCidr = ref("192.168.1.0/24");
const community = ref("");
const report = ref<NetworkReport | null>(null);
const isScanning = ref(false);
const searchQuery = ref("");

const selectedDevice = ref<Device | null>(null);

const filteredDevices = computed(() => {
  if (!report.value) return [];

  return report.value.devices
    .filter(d => d.status === "Online") // 🔥 ONLY ONLINE
    .filter(d => 
      d.ip.includes(searchQuery.value) || 
      d.mac.toLowerCase().includes(searchQuery.value.toLowerCase())
    );
});

function openDevice(device: Device) {
  selectedDevice.value = device;
}

function closeDevice() {
  selectedDevice.value = null;
}


async function startAudit() {
  if (isScanning.value) return;
  isScanning.value = true;
  report.value = null;

  try {
    const result = await invoke<string>("execute_enterprise_audit", { 
      args: { target: targetCidr.value, community: community.value || null } 
    });
    console.log("RAW RESULT:", result);
    report.value = JSON.parse(result);
  } catch (error) {
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
  a.download = `empire-audit-${Date.now()}.csv`;
  a.click();
};
</script>

<template>
  <main class="relative min-h-screen text-slate-300 font-sans overflow-x-hidden">
    
    <div class="fixed inset-0 flex items-center justify-center pointer-events-none select-none z-0 overflow-hidden">
      <h1 class="text-[15vw] font-black text-white/[0.02] uppercase tracking-[2rem] leading-none whitespace-nowrap rotate-[-12deg]">
        EMPIRE NETWORK TOOL
      </h1>
    </div>

    <div class="fixed top-[-10%] left-[-10%] w-[40%] h-[40%] bg-blue-600/10 blur-[120px] rounded-full"></div>
    <div class="fixed bottom-[-10%] right-[-10%] w-[30%] h-[30%] bg-indigo-600/10 blur-[120px] rounded-full"></div>

    <nav class="relative z-10 border-b border-white/5 bg-black/40 backdrop-blur-xl sticky top-0">
      <div class="max-w-7xl mx-auto px-8 h-20 flex items-center justify-between">
        <div class="flex items-center gap-4">
          <div class="relative group">
            <div class="absolute -inset-1 bg-gradient-to-r from-blue-600 to-cyan-500 rounded-lg blur opacity-25 group-hover:opacity-50 transition duration-1000"></div>
            <div class="relative w-10 h-10 bg-black rounded-lg flex items-center justify-center border border-white/10">
              <span class="font-black text-blue-500 text-sm">EN</span>
            </div>
          </div>
          <div>
            <h1 class="font-black tracking-[0.2em] text-lg uppercase text-white">EMPIRE <span class="text-blue-500 font-light">MAINTAINER</span></h1>
            <p class="text-[10px] text-slate-500 font-mono tracking-widest uppercase">Enterprise Auditor v3.0</p>
          </div>
        </div>
        <div class="flex items-center gap-6">
          <div v-if="isScanning" class="flex items-center gap-2 px-3 py-1 bg-blue-500/10 border border-blue-500/20 rounded-full">
            <span class="w-2 h-2 bg-blue-500 rounded-full animate-ping"></span>
            <span class="text-[10px] font-bold text-blue-400 uppercase tracking-tighter">Probing Subnet...</span>
          </div>
          <button v-if="report" @click="exportCSV" class="text-[10px] font-bold uppercase tracking-widest bg-white/5 hover:bg-white/10 px-6 py-2.5 rounded-full border border-white/10 transition-all active:scale-95">
            Export Intel
          </button>
        </div>
      </div>
    </nav>

    <div class="relative z-10 max-w-7xl mx-auto px-8 py-12">
      <div class="grid grid-cols-12 gap-10">
        
        <aside class="col-span-12 lg:col-span-4 space-y-8">
          <section class="bg-[#0a0a0c]/80 backdrop-blur-md p-8 rounded-3xl border border-white/10 shadow-2xl relative overflow-hidden group">
            <div class="absolute top-0 left-0 w-full h-1 bg-gradient-to-r from-transparent via-blue-500/40 to-transparent"></div>
            <h2 class="text-[10px] font-black uppercase tracking-[0.3em] text-slate-500 mb-8 flex items-center gap-2">
              <span class="w-4 h-[1px] bg-slate-700"></span> System Configuration
            </h2>
            
            <div class="space-y-6">
              <div class="relative">
                <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-3 ml-1">Subnet Range (CIDR)</label>
                <input v-model="targetCidr" 
                  class="w-full bg-black/40 border border-white/5 rounded-2xl px-5 py-4 text-sm focus:border-blue-500/50 focus:ring-4 focus:ring-blue-500/5 outline-none transition-all font-mono text-blue-400"
                  placeholder="10.0.0.0/24" />
              </div>

              <div class="relative">
                <label class="block text-[10px] font-bold text-slate-500 uppercase tracking-widest mb-3 ml-1">SNMP Community Key</label>
                <input v-model="community" type="password"
                  class="w-full bg-black/40 border border-white/5 rounded-2xl px-5 py-4 text-sm focus:border-blue-500/50 focus:ring-4 focus:ring-blue-500/5 outline-none transition-all font-mono"
                  placeholder="Secure String" />
              </div>

              <button @click="startAudit" :disabled="isScanning"
                class="group relative w-full py-5 rounded-2xl font-black text-[11px] tracking-[0.2em] uppercase transition-all overflow-hidden"
                :class="isScanning ? 'bg-slate-800 text-slate-600' : 'bg-white text-black hover:bg-blue-500 hover:text-white'">
                <span class="relative z-10">{{ isScanning ? 'Synchronizing...' : 'Initialize Deep Scan' }}</span>
              </button>
            </div>
          </section>

          <div v-if="report" class="p-6 rounded-3xl border border-white/5 bg-gradient-to-br from-blue-500/[0.03] to-transparent animate-empire-in">
            <div class="flex justify-between items-start mb-4">
              <h3 class="text-[10px] font-black uppercase tracking-widest text-blue-500">Node Performance</h3>
              <span class="text-[10px] font-mono text-slate-600">{{ report.timestamp }}</span>
            </div>
            <div class="flex items-baseline gap-2">
              <span class="text-3xl font-black text-white italic">{{ filteredDevices.length }}</span>
              <span class="text-[10px] text-slate-500 uppercase font-bold tracking-widest">Active Endpoints</span>
            </div>
          </div>
        </aside>

        <div class="col-span-12 lg:col-span-8 space-y-8">
          <div v-if="!report && !isScanning" class="group h-[400px] flex flex-col items-center justify-center border border-white/5 rounded-[2.5rem] bg-white/[0.01] transition-all hover:bg-white/[0.02] border-dashed">
             <div class="w-16 h-16 mb-6 rounded-full border border-white/10 flex items-center justify-center group-hover:scale-110 transition-transform duration-500">
               <div class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
             </div>
             <p class="text-[11px] font-black tracking-[0.3em] uppercase text-slate-600">Awaiting Command Input</p>
          </div>

          <div v-if="report" class="space-y-6 animate-empire-in">
            <div class="flex items-center justify-between gap-6">
              <div class="relative flex-1">
                <span class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-600">
                  <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg>
                </span>
                <input v-model="searchQuery" 
                  class="bg-[#0a0a0c] border border-white/10 rounded-2xl pl-12 pr-6 py-4 text-sm w-full focus:border-blue-500/50 outline-none transition-all placeholder:text-slate-700"
                  placeholder="Query Registry by IP or MAC..." />
              </div>
              <div class="px-5 py-4 bg-white/5 rounded-2xl border border-white/10 text-[10px] font-black uppercase tracking-widest text-slate-500">
                L2 Protocol: {{ report.scanMethod }}
              </div>
            </div>

            <div class="bg-[#0a0a0c] border border-white/10 rounded-[2.5rem] overflow-hidden shadow-3xl backdrop-blur-sm">
              <table class="w-full text-left">
                <thead>
                      
                    <tr >
                        
                    <th class="px-8 py-6 font-black">Endpoint Address</th>
                    <th class="px-8 py-6 font-black">Hardware Identity</th>
                    <th class="px-8 py-6 font-black text-right">Verification</th>
                    
                    <th class="px-8 py-6 font-black">TTL / OS</th>
                    <th class="px-8 py-6 font-black">Subnet</th>
                  </tr>
                  
                </thead>
                <tbody class="divide-y divide-white/5">
                  <tr 
                    v-for="device in filteredDevices" 
                    :key="device.ip" 
                    @click="openDevice(device)" 
                    class="cursor-pointer hover:bg-blue-500/[0.05] transition"
                  >
                    <td class="px-8 py-6">
                      <div class="flex items-center gap-3">
                        <div class="w-1.5 h-1.5 rounded-full bg-blue-500 shadow-[0_0_10px_rgba(59,130,246,0.5)]"></div>
                        <span class="text-sm font-mono font-bold text-white group-hover:text-blue-400 transition-colors">{{ device.ip }}</span>
                        <span class="text-xs text-slate-600">(view)</span>
                      </div>
                    </td>
                    <td class="px-8 py-6 text-xs font-mono text-slate-500 group-hover:text-slate-300">{{ device.mac }}</td>
                    <td class="px-8 py-6 text-right">
                      
                      <span 
                        class="inline-flex items-center px-4 py-1.5 rounded-full text-[9px] font-black uppercase tracking-widest border"
                        :class="device.status === 'Online' 
                          ? 'bg-emerald-500/5 text-emerald-500 border-emerald-500/10' 
                          : 'bg-red-500/5 text-red-400 border-red-500/10'"
                      >
                        {{ device.status }}
                      </span>

                    </td>
                    
                    
                    <td class="px-8 py-6 text-xs font-mono">
                      {{ device.ttl }} / {{ device.os }}
                    </td>

                    <td class="px-8 py-6 text-xs">
                      <span :class="device.subnetMatch ? 'text-green-400' : 'text-red-400'">
                        {{ device.subnetMatch ? 'Same' : 'Mismatch' }}
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
  
  <div 
    v-if="selectedDevice" 
    class="fixed top-0 right-0 h-full w-[400px] bg-[#0a0a0c] border-l border-white/10 shadow-2xl z-50 p-8 transition-transform"
  >
    <h2 class="text-lg font-bold mb-6 text-blue-400">Device Intelligence</h2>

    <div class="space-y-3 text-sm font-mono">
      <p><b>IP:</b> {{ selectedDevice.ip }}</p>
      <p><b>MAC:</b> {{ selectedDevice.mac }}</p>
      <p><b>Status:</b> {{ selectedDevice.status }}</p>
      <p><b>TTL:</b> {{ selectedDevice.ttl }}</p>
      <p><b>OS Guess:</b> {{ selectedDevice.os }}</p>
      <p>
        <b>Subnet Match:</b> 
        <span :class="selectedDevice.subnetMatch ? 'text-green-400' : 'text-red-400'">
          {{ selectedDevice.subnetMatch ? 'Yes' : 'Mismatch' }}
        </span>
      </p>
    </div>

    <button 
      @click="closeDevice"
      class="mt-8 w-full py-3 bg-white text-black rounded-xl font-bold"
    >
      Close
    </button>
  </div>

  </div>


</template>