<script setup lang="ts">
import { ref } from 'vue';
import { invoke } from "@tauri-apps/api/core";

interface Report {
  activeSubnets: string[];
  alienIPs: string[];
  performance: string;
  suggestions: string[];
}

const subnet = ref("192.168.1.0/24");
const report = ref<Report | null>(null);
const loading = ref(false);

const startScan = async () => {
  loading.value = true;
  try {
    const rawJson = await invoke<string>("run_network_scan", { subnet: subnet.value });
    report.value = JSON.parse(rawJson);
  } catch (err) {
    alert("Scan failed: " + err);
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <main class="p-10 bg-slate-900 min-h-screen text-slate-100">
    <div class="max-w-4xl mx-auto">
      <h1 class="text-4xl font-black mb-2 text-blue-400">LAN MAINTAINER</h1>
      <p class="text-slate-400 mb-8 font-mono">Enterprise Network Auditor v2.0</p>

      <div class="flex gap-4 mb-10">
        <input v-model="subnet" class="bg-slate-800 border border-slate-700 px-4 py-2 rounded flex-1 font-mono" />
        <button @click="startScan" :disabled="loading" 
          class="bg-blue-600 hover:bg-blue-500 disabled:opacity-50 px-8 py-2 rounded font-bold uppercase tracking-widest transition-all">
          {{ loading ? 'Analyzing...' : 'Execute Audit' }}
        </button>
      </div>

      <div v-if="report" class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div :class="['p-6 rounded-lg border-l-4', report.performance === 'Optimal' ? 'bg-emerald-900/20 border-emerald-500' : 'bg-red-900/20 border-red-500']">
          <h2 class="text-sm font-bold text-slate-400 uppercase mb-1">Performance</h2>
          <p class="text-2xl font-bold">{{ report.performance }}</p>
          <div class="mt-4 space-y-1">
            <p v-for="s in report.suggestions" :key="s" class="text-sm text-slate-300">💡 {{ s }}</p>
          </div>
        </div>

        <div class="bg-slate-800 p-6 rounded-lg border border-slate-700">
          <h2 class="text-sm font-bold text-slate-400 uppercase mb-1">Security Audit</h2>
          <p class="text-lg">Alien IPs: <span :class="report.alienIPs.length > 0 ? 'text-red-400' : 'text-emerald-400'">{{ report.alienIPs.length }} Detected</span></p>
          <ul class="mt-4 font-mono text-xs text-red-300 space-y-1">
            <li v-for="ip in report.alienIPs" :key="ip">{{ ip }} (Unknown Device)</li>
          </ul>
        </div>
      </div>
    </div>
  </main>
</template>