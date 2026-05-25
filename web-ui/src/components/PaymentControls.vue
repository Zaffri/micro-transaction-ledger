<script setup lang="ts">

interface Props {
  makePayment: (sender: number, amount: number) => void
}

const props = defineProps<Props>();
const senderId = defineModel<number>('senderId', { default: -1 });
const amount = defineModel<number>('amount', { default: 0 });
</script>

<template>
  <div class="px-4 pt-4 bg-slate-100/60 shrink-0">
    <div class="flex flex-wrap items-center justify-between gap-4 rounded-2xl border border-slate-200/80 bg-white p-4 shadow-xs">

      <div class="flex flex-1 flex-wrap items-center gap-3 min-w-[280px]">
        <div class="relative w-full sm:w-65">
          <select v-model.number="senderId" class="w-full appearance-none rounded-xl border border-slate-200 bg-slate-50 px-3.5 py-2 text-sm font-medium text-slate-700 outline-none transition-all hover:bg-slate-100/70 focus:border-indigo-500 focus:bg-white focus:ring-2 focus:ring-indigo-500/10">
            <option value="-1">Select transer direction</option>
            <option value="1">From John's to Jane's account</option>
            <option value="2">From Jane's to John's account</option>
          </select>

          <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3.5 text-slate-400">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="h-4 w-4">
              <path fill-rule="evenodd" d="M5.22 8.22a.75 .75 0 0 1 1.06 0L10 11.94l3.72-3.72a.75 .75 0 1 1 1.06 1.06l-4.25 4.25a.75 .75 0 0 1-1.06 0L5.22 9.28a.75 .75 0 0 1 0-1.06Z" clip-rule="evenodd" />
            </svg>
          </div>
        </div>

        <div class="relative w-full sm:w-56">
          <div class="pointer-events-none absolute inset-y-0 left-0 flex items-center pl-3.5">
            <span class="text-sm font-medium text-slate-400">£</span>
          </div>

          <input
            v-model="amount"
            type="number" 
            placeholder="0.00" 
            step="0.01"
            class="w-full rounded-xl border border-slate-200 bg-slate-50 py-2 pl-8 pr-16 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-indigo-500 focus:bg-white focus:ring-2 focus:ring-indigo-500/10"
          />

          <div class="pointer-events-none absolute inset-y-0 right-0 flex items-center pr-3.5">
            <span class="text-xs font-semibold tracking-wider text-slate-400">GBP</span>
          </div>
        </div>

        <button @click="makePayment(senderId, amount)" class="inline-flex w-full sm:w-auto items-center justify-center gap-2 rounded-xl bg-blue-700 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-blue-900 shadow-xs">
          <span>Send</span>
        </button>
      </div>
    </div>
  </div>
</template>
