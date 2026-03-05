<script setup lang="ts">
import { ref } from 'vue';
import { invoke } from "@tauri-apps/api/core";

const target = ref("192.168.1.1");
const community = ref(""); // Empty = Unmanaged mode
const results = ref<any>(null);
const isRunning = ref(false);

async function runAudit() {
  isRunning.value = true;
  try {
    const raw = await invoke("execute_enterprise_audit", { 
      args: { target: target.value, community: community.value || null } 
    });
    results.value = JSON.parse(raw as string);
  } catch (e) {
    console.error(e);
  } finally {
    isRunning.value = false;
  }
}
</script>

<template>
  <div class="p-8 bg-black text-white min-h-screen font-sans">
    <header class="border-b border-gray-800 pb-4 mb-8">
      <h1 class="text-2xl font-bold tracking-tighter">LAN_MAINTAINER <span class="text-blue-500">PRO</span></h1>
    </header>

    <div class="grid grid-cols-12 gap-6">
      <div class="col-span-4 space-y-4">
        <div class="bg-zinc-900 p-4 rounded border border-zinc-800">
          <label class="block text-xs uppercase text-zinc-500 mb-2">Network Target (IP/Subnet)</label>
          <input v-model="target" class="w-full bg-zinc-800 p-2 rounded text-sm outline-none border border-transparent focus:border-blue-500" />
          
          <label class="block text-xs uppercase text-zinc-500 mt-4 mb-2">SNMP Community (Optional)</label>
          <input v-model="community" placeholder="Leave empty for unmanaged" class="w-full bg-zinc-800 p-2 rounded text-sm outline-none border border-transparent focus:border-blue-500" />
          
          <button @click="runAudit" :disabled="isRunning" class="w-full mt-6 bg-blue-600 hover:bg-blue-700 p-3 rounded font-bold transition-colors">
            {{ isRunning ? 'PROBING NETWORK...' : 'START AUDIT' }}
          </button>
        </div>
      </div>

      <div class="col-span-8">
        <div v-if="results" class="bg-zinc-900 rounded border border-zinc-800 overflow-hidden">
          <div class="p-4 bg-zinc-800 flex justify-between items-center">
            <span class="text-xs font-mono text-zinc-400">METHOD: {{ results.scanMethod }}</span>
            <span class="text-xs font-mono text-emerald-400">STATUS: {{ results.performance }}</span>
          </div>
          <table class="w-full text-left text-sm">
            <thead class="text-zinc-500 border-b border-zinc-800">
              <tr>
                <th class="p-4">IP Address</th>
                <th class="p-4">Hardware (MAC)</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="dev in results.devices" :key="dev.ip" class="border-b border-zinc-800/50 hover:bg-zinc-800/30">
                <td class="p-4 font-mono">{{ dev.ip }}</td>
                <td class="p-4 text-zinc-400">{{ dev.mac }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>