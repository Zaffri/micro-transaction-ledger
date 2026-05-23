<script setup lang="ts">
import AccountPane from './components/AccountPane.vue';
import PaymentControls from './components/PaymentControls.vue';
import Topbar from './components/Topbar.vue';

</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-slate-50 font-sans text-slate-700 antialiased overflow-hidden">
    <Topbar />

    <PaymentControls />

    <main class="flex min-h-0 flex-1 px-4 pb-4 pt-3 gap-4 bg-slate-100/60">
      <AccountPane account-name="John" :balance="1000" />
      <AccountPane account-name="Jane" :balance="3000" />
    </main>
  </div>
</template>

<style scoped>
input[type="number"]::-webkit-inner-spin-button,
input[type="number"]::-webkit-outer-spin-button {
  -webkit-appearance: none;
  margin: 0;
}
</style>
