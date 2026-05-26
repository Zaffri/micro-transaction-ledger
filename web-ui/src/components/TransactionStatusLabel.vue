<script setup lang="ts">
import { computed } from 'vue';


type Status = 'pending' | 'settled' | 'rejected_fraud';

interface Props {
  status: Status,
  amountInPennies: number;
}

const props = defineProps<Props>()

const labelStyles = computed<string>(() => {
  const baseStyles = 'rounded-md border px-2 py-0.5 text-[12px] font-medium uppercase ';
  const compensatingStyles = 'text-slate bg-slate-200 border-slate-300';

  const statusColours: Record<Status, string> = {
    pending: 'text-white bg-yellow-600 border-black-100',
    settled: 'text-white bg-green-700 border-black-100',
    rejected_fraud: 'text-white bg-red-700 border-black-100',
  };

  const additionalStyles = (isCompensatingTransaction.value) ? compensatingStyles : statusColours[props.status] ?? '';

  return baseStyles + additionalStyles;
});

const isCompensatingTransaction = computed(() => {
  return props.status === "rejected_fraud" && props.amountInPennies > 0;
});

const labelText = computed<string>(() => {
  if (isCompensatingTransaction.value) return 'reversal';

  const mapping: Record<Status, string> = {
    pending: "pending",
    settled: "settled",
    rejected_fraud: "fraud",
  };

  return mapping[props.status] ?? '';
});

</script>

<template>
  <div class="w-20">
    <span :class="labelStyles">
      {{ labelText }}
    </span>
  </div>
</template>
