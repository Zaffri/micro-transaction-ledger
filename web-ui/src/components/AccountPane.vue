<script setup lang="ts">
import { computed } from 'vue';
import AccountTransaction from './AccountTransaction.vue';

interface Props {
  accountName: string
  balance: number
};

const props = defineProps<Props>();

const accountTitle = computed<string>(() => {
  if (!props.accountName) return '';
  return `${props.accountName}'s Account`
})

</script>

<template>
  <section class="flex flex-col flex-1 overflow-y-auto bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 custom-scrollbar">
    <div>
      <div class="flex w-full items-baseline justify-between border-b border-slate-100 pb-3">
        <h2 class="text-2xl font-bold text-slate-900 mb-4">{{ accountTitle }}</h2>

        <span class="text-lg font-bold text-slate-600">
          Balance: £{{ props.balance }}
        </span>
      </div>

      <AccountTransaction type="credit" :amount="10", status="pending" :date="new Date('2026-05-22')" />
      <AccountTransaction type="credit" :amount="1000", status="settled" :date="new Date('2026-05-22')" />
      <AccountTransaction type="debit" :amount="150", status="rejected_fraud" :date="new Date('2026-05-22')" />
      <AccountTransaction type="debit" :amount="50", status="settled" :date="new Date('2026-05-22')" />
    </div>
  </section>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 8px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--color-slate-200, #e2e8f0);
  border-radius: 9999px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: var(--color-slate-300, #cbd5e1);
}

</style>
