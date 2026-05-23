<script setup lang="ts">
import { computed } from 'vue';


type Status = 'pending' | 'settled' | 'rejected_fraud';

interface Props {
  status: Status,
}

const props = defineProps<Props>()

const labelStyles = computed<string>(() => {
  const baseStyles = 'rounded-md border px-2 py-0.5 text-[12px] font-medium uppercase ';

  const statusColours: Record<Status, string> = {
    pending: 'text-white bg-yellow-600 border-black-100',
    settled: 'text-white bg-green-700 border-black-100',
    rejected_fraud: 'text-white bg-red-700 border-black-100',
  };

  return baseStyles + (statusColours[props.status] ?? '');
});

const labelText = computed<string>(() => {
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
