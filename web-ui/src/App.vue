<script setup lang="ts">
import { onMounted, ref } from 'vue';
import AccountPane from './components/AccountPane.vue';
import PaymentControls from './components/PaymentControls.vue';
import Topbar from './components/Topbar.vue';
import { getAccounts, makePayment, type Account, type StatementLine } from './data/api';

const isLoading = ref(true);
const accounts = ref<Account[]>([]);
const polling = ref<boolean>(false);
const pollingTimer = ref<number | undefined>();
const errorMessage = ref('')

const getData = async () => {
  const results = await getAccounts();

  if ('error' in results) {
    // TODO: SET AND SHOW ERROR - errorMessage
    console.log(results);
    return;
  }

  accounts.value = results.data;
  isLoading.value = false;
};

const startPolling = () => {
  if (polling.value) return;
  polling.value = true;

  pollingTimer.value = setInterval(() => {
    getData();
  }, 10000);
};

const stopPolling = () => {
  if (!polling.value) return;
  polling.value = false;

  clearInterval(pollingTimer.value)
  pollingTimer.value = undefined;
};

const handlePayment = async (senderId: number, amount: number) => {
  const positionInAccounts = accounts.value.findIndex(account => account.ID === senderId);
  if (positionInAccounts === -1) return;

  stopPolling();

  const result = await makePayment(senderId, amount);
  const otherPartyIndex = positionInAccounts === 1 ? 2 : 1;
  const otherPartyName = (accounts.value[otherPartyIndex]) ? accounts.value[otherPartyIndex].AccountHolderName : '';

  // TODO: handle error

  if (accounts.value[positionInAccounts]) {
    const optimisticUpdate: StatementLine = {
      Status: 'pending',
      AmountInPennies: amount,
      OtherPartyName: otherPartyName,
      CreatedAt: new Date().toDateString()
    }

    accounts.value[positionInAccounts].Statement.unshift(optimisticUpdate)
  }

  console.log(result);
  startPolling();
};

onMounted(() => {
  getData();
  startPolling();
});

</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-slate-50 font-sans text-slate-700 antialiased overflow-hidden">
    <Topbar />

    <PaymentControls :make-payment="handlePayment" />

    <main class="flex min-h-0 flex-1 px-4 pb-4 pt-3 gap-4 bg-slate-100/60">
      <AccountPane v-for="account in accounts" :is-loading="isLoading" :account="account" />
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
