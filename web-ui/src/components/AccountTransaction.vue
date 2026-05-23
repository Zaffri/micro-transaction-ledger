<script setup lang="ts">
import TransactionStatusLabel from './TransactionStatusLabel.vue';


interface Props {
  type: 'credit' | 'debit',
  amount: number,
  status: 'pending' | 'settled' | 'rejected_fraud',
  date: Date
}

const props = defineProps<Props>()

</script>

<template>
  <div
    class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 rounded-xl border border-slate-100 bg-slate-50/50 p-4 transition-all hover:bg-slate-50 hover:border-slate-200/60 mb-3"
  >
    <div class="flex flex-1 items-center gap-4 min-w-0">
    
      <div class="shrink-0 text-xs font-semibold tracking-wide text-slate-400 w-25">
        {{ props.date.toDateString() }}
      </div>

      <div class="flex flex-wrap items-center gap-2 min-w-0">
        <TransactionStatusLabel :status="props.status" />

        <span class="text-sm font-medium text-slate-600 truncate max-w-[200px] sm:max-w-xs">
          Transfer from X
        </span>
      </div>

    </div>

    <div class="flex shrink-0 items-center justify-end text-sm font-bold sm:text-right">
      <div class="shrink-0 mr-5">
        <span 
          :class="[
            'inline-flex items-center gap-1 rounded-lg px-2.5 py-1 text-xs font-bold tracking-wide uppercase shadow-xs',
            props.type === 'credit' 
              ? 'bg-emerald-50 text-emerald-700 border border-emerald-200/60' 
              : 'bg-rose-50 text-rose-700 border border-rose-200/60'
          ]"
        >
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor" class="h-3.5 w-3.5">
            <path v-if="props.type === 'credit'" fill-rule="evenodd" d="M8 12.5a.75.75 0 0 1-.75-.75V5.56L4.78 8.03a.75.75 0 0 1-1.06-1.06l3.75-3.75a.75.75 0 0 1 1.06 0l3.75 3.75a.75.75 0 1 1-1.06 1.06L8.75 5.56V11.75a.75.75 0 0 1-.75.75Z" clip-rule="evenodd" />
            <path v-else fill-rule="evenodd" d="M8 3.5a.75.75 0 0 1 .75.75v6.19l2.47-2.47a.75.75 0 1 1 1.06 1.06l-3.75 3.75a.75.75 0 0 1-1.06 0l-3.75-3.75a.75.75 0 1 1 1.06-1.06l2.47 2.47V4.25A.75.75 0 0 1 8 3.5Z" clip-rule="evenodd" />
          </svg>
          {{ props.type === 'credit' ? 'in' : 'out' }}
        </span>
      </div>

      <span class="text-slate-600">
        {{ props.type === 'credit' ? '+' : '-' }}£{{props.amount}}
      </span>
    </div>

  </div>
</template>