<script setup lang="ts">
import { computed } from 'vue';
import AccountTransaction from './AccountTransaction.vue';
import type { Account } from '@/data/api';

interface Props {
  isLoading: boolean;
  account: Account
};

const props = defineProps<Props>();

const accountTitle = computed<string>(() => {
  if (!props.account) return '';
  return `${props.account.AccountHolderName}'s Account`
})

const balance = computed<number | string>(() => {
  if (!props.account) return '--';
  // TODO: formatting...
  return props.account.BalanceInPennies;
});

</script>

<template>
  <section class="flex flex-col flex-1 overflow-y-auto bg-white rounded-2xl border border-slate-200/80 shadow-xs p-6 custom-scrollbar">
      <div v-if="!props.isLoading">
        <div class="flex w-full items-baseline justify-between border-b border-slate-100 pb-3">
          <h2 class="text-2xl font-bold text-slate-900 mb-4">{{ accountTitle }}</h2>

          <span class="text-lg font-bold text-slate-600">
            Balance: {{ balance }}
          </span>
        </div>

        <TransitionGroup tag="div" name="statement-line">
          <AccountTransaction
            v-for="transaction in account.Statement"
            :key="transaction.ID"
            :statement="transaction"
          />
        </TransitionGroup>
      </div>

      <div v-else>
        <div class="flex w-full items-baseline justify-between border-b border-slate-100 pb-3">
          <h2 class="text-2xl font-bold text-slate-700 mb-4">Loading...</h2>
  
          <span class="text-lg font-bold text-slate-500">
            Loading...
          </span>
        </div>

        <div class="text-slate-500 pt-5">Loading...</div>
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

:deep(.statement-line-enter-active,
.statement-line-leave-active) {
  transition: all 0.5s ease;
}

:deep(.statement-line-enter-from,
.statement-line-leave-to) {
  opacity: 0;
  transform: translateX(30px);
}
</style>
