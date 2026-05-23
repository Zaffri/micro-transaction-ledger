<script setup lang="ts">
import { onMounted, ref } from 'vue';
import AccountPane from './components/AccountPane.vue';
import PaymentControls from './components/PaymentControls.vue';
import Topbar from './components/Topbar.vue';

const isLoading = ref(true);
const accounts = ref<Account[]>([]);
const errorMessage = ref('')

const API_GATEWAY = 'http://localhost:8080'
const ACCOUNTS_ENDPOINT = '/accounts'

export interface StatementLine {
  Status: 'settled' | 'rejected_fraud' | 'pending';
  AmountInPennies: number;
  OtherPartyName: string;
  CreatedAt: string;
}

export interface Account {
  ID: number;
  BalanceInPennies: number;
  AccountHolderName: string;
  CreatedAt: string;
  UpdatedAt: string;
  Statement: StatementLine[];
}

const getAccounts = async (): Promise<{ data: Account[] } | { error: string }> => {
  try {
    const accountRequests = [
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/1`).then(res => res.json() as Promise<Account>),
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/2`).then(res => res.json() as Promise<Account>),
    ];

    const statementRequests = [
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/1/statement`).then(res => res.json() as Promise<StatementLine[]>),
      fetch(`${API_GATEWAY}${ACCOUNTS_ENDPOINT}/2/statement`).then(res => res.json() as Promise<StatementLine[]>),
    ];
  
    const [accountOne, accountTwo] = await Promise.all(accountRequests);
    const [accountOneStatement, accountTwoStatement] = await Promise.all(statementRequests);
    
    const preparedData: Account[] = [];

    if (accountOne && accountOneStatement) {
      accountOne.Statement = accountOneStatement
      preparedData[0] = accountOne;
    }

    if (accountTwo && accountTwoStatement) {
      accountTwo.Statement = accountTwoStatement
      preparedData[1] = accountTwo;
    }

    return { data: preparedData };
  } catch (err) {
    console.log(err);
    return { error: 'Failed to fetch account data' };
  }
};

onMounted(() => {
  getAccounts()
    .then((results) => {
      if ('error' in results) {
        // TODO: SET AND SHOW ERROR - errorMessage
        console.log(results);
        return;
      }

      accounts.value = results.data;
      isLoading.value = false;
      console.log("Setting loading false", isLoading.value);
    })
    .catch(() => {
      // TODO: set error and stuff
    });
});

</script>

<template>
  <div class="flex h-screen w-screen flex-col bg-slate-50 font-sans text-slate-700 antialiased overflow-hidden">
    <Topbar />

    <PaymentControls />

    <main class="flex min-h-0 flex-1 px-4 pb-4 pt-3 gap-4 bg-slate-100/60">
      <!-- <AccountPane :is-loading="isLoading" :account-name="accounts[1]?.AccountHolderName" :balance="1000" /> -->
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
