<script setup lang="ts">
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import AppInput from '../../components/ui/AppInput.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { computed, onMounted, ref, watch } from 'vue';
import { WalletCards } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import type { CurrencyUnit } from '../../utils/currency';
import { formatMoney, formatMoneyInput, groupInteger, parseMoneyInput } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
import { accountingApi, type FinancialAccountRecord, type PaymentRecord } from '../../api/accounting';
import { ordersApi } from '../../api/orders';
import SelectField from '../../components/ui/SelectField.vue';
const props = defineProps<{ order: any; currencyUnit: CurrencyUnit }>();
const emit = defineEmits<{ notify: [message: string]; saved: [order: any] }>();
const financial = ref<FinancialAccountRecord[]>([]);
const payments = ref<PaymentRecord[]>([]);
const amount = ref('');
const account = ref('FIN-CASH');
const method = ref('cash');
const loading = ref(false);
const money = (n: number) => formatMoney(n, props.currencyUnit);
const allPaymentMethods = [
  { label: 'Cash', value: 'cash' },
  { label: 'Bank transfer', value: 'bank_transfer' },
  { label: 'Card', value: 'card' },
  { label: 'Check', value: 'check' },
];
const selectedFinancialAccount = computed(() =>
  financial.value.find((value) => value.id === account.value),
);
const paymentMethodOptions = computed(() =>
  selectedFinancialAccount.value?.type === 'cash'
    ? allPaymentMethods.filter((value) => value.value === 'cash')
    : selectedFinancialAccount.value?.type === 'bank'
      ? allPaymentMethods.filter((value) => value.value !== 'cash')
      : allPaymentMethods,
);
function formatAmountWhileTyping(value: string) {
  const raw = value.replaceAll(',', '');
  if (raw === '' || raw === '-') {
    amount.value = raw;
    return;
  }
  if (!/^-?\d*(?:\.\d*)?$/.test(raw)) return;
  const [whole, fraction] = raw.split('.');
  const groupedWhole = groupInteger(whole || (raw.startsWith('-') ? '-0' : '0'));
  amount.value = fraction === undefined ? groupedWhole : `${groupedWhole}.${fraction.slice(0, 1)}`;
}
watch(
  paymentMethodOptions,
  (options) => {
    if (!options.some((value) => value.value === method.value)) {
      method.value = options[0]?.value ?? '';
    }
  },
  { immediate: true },
);
async function load() {
  loading.value = true;
  try {
    financial.value = await accountingApi.financialAccounts();
    if (!financial.value.some((value) => value.id === account.value)) {
      account.value = financial.value[0]?.id ?? '';
    }
    payments.value = (await accountingApi.payments()).filter((p) =>
      p.allocations.some((a) => a.targetType === 'order' && a.targetId === props.order.id),
    );
  } catch (e) {
    emit('notify', String(e));
  } finally {
    loading.value = false;
  }
}
async function refreshOrder() {
  emit('saved', await ordersApi.get(props.order.id));
}
async function post() {
return runAction(async () => {
  try {
    const n = parseMoneyInput(amount.value, props.currencyUnit);
    if (n === null || n <= 0) throw new Error('Enter a positive amount');
    await accountingApi.createPayment({
      direction: 'incoming',
      method: method.value,
      financialAccountId: account.value,
      customerId: props.order.customerId,
      amountRial: n,
      allocations: [{ targetType: 'order', targetId: props.order.id, amountRial: n }],
    });
    amount.value = '';
    await load();
    await refreshOrder();
    emit('notify', 'Order payment posted');
  } catch (e) {
reportError(e);
  }

});
}
async function reverse(p: PaymentRecord) {
return runAction(async () => {
  try {
    await accountingApi.reversePayment(p.id);
    await load();
    await refreshOrder();
    emit('notify', 'Payment reversed');
  } catch (e) {
reportError(e);
  }

});
}
onMounted(load);
</script>
<template>
  <div class="min-w-0 space-y-3">
    <AppPanel
      title="Payment status"
      subtitle="Paid and remaining amounts are derived by the backend."
      ><div class="grid min-w-0 gap-3 sm:grid-cols-3">
        <div class="rounded-box border border-base-300 bg-base-200/55 p-3">
          <span class="block text-xs text-base-content/55">Order total</span>
          <strong class="mt-1 block text-base tabular-nums">{{ money(order.totalRial) }}</strong>
        </div>
        <div class="rounded-box border border-base-300 bg-base-200/55 p-3">
          <span class="block text-xs text-base-content/55">Paid</span>
          <strong class="mt-1 block text-base tabular-nums text-success">{{
            money(order.paidRial || 0)
          }}</strong>
        </div>
        <div class="rounded-box border border-primary/25 bg-primary/5 p-3">
          <span class="block text-xs text-base-content/55">Remaining</span>
          <strong class="mt-1 block text-base tabular-nums text-primary">{{
            money(order.remainingRial ?? order.totalRial)
          }}</strong>
        </div>
      </div>
      <form @submit.prevent="post" class="mt-4 grid min-w-0 gap-3 border-t border-base-300 pt-4 sm:grid-cols-2">
        <SelectField
          v-model="account"
          label="Financial account"
          :options="financial.map((value) => ({ label: value.name, value: value.id }))"
        /><SelectField
          v-model="method"
          label="Method"
          :options="paymentMethodOptions"
        /><FormField class="sm:col-span-2" :label="`Amount (${props.currencyUnit})`"><AppInput
          class="input w-full min-w-0"
          v-model="amount"
          :money="props.currencyUnit"
          :placeholder="`Amount (${props.currencyUnit})`"
          inputmode="numeric"
          @update:model-value="formatAmountWhileTyping"
          @blur="
            amount = formatMoneyInput(
              parseMoneyInput(amount, props.currencyUnit) || 0,
              props.currencyUnit,
            )
          "
        /></FormField><button class="btn btn-primary w-fit gap-2" type="submit" :disabled="busy">
          <WalletCards :size="15" aria-hidden="true" />
          {{ busy ? 'Posting…' : 'Post payment' }}
        </button>
      </form></AppPanel
    ><AppPanel
      title="Order payment history"
      subtitle="Reversals preserve the original payment and journal entry."
      ><div
        v-if="!payments.length"
        class="flex min-h-28 flex-col items-center justify-center gap-2 rounded-box border border-dashed border-base-300 bg-base-200/35 p-4 text-center"
      >
        <WalletCards :size="22" class="text-base-content/45" aria-hidden="true" />
        <span class="text-sm text-base-content/65">No allocated payments for this order.</span>
      </div>
      <div
        v-for="p in payments"
        :key="p.id"
        class="flex min-w-0 flex-wrap items-center justify-between gap-3 rounded-box border border-base-300 bg-base-200/35 p-3"
      >
        <span
          ><strong class="block text-sm">{{ p.paymentNumber }}</strong
          ><small class="block text-xs leading-5 text-base-content/60"
            >{{ money(p.amountRial) }} · {{ formatDateTime(p.postedAt) }}</small
          ></span
        ><span
          class="flex items-center gap-2"
          ><StatusBadge
            :label="p.status"
            :tone="p.status === 'posted' ? 'green' : 'slate'"
          /><button class="btn btn-ghost btn-sm" v-if="p.status === 'posted'" @click="reverse(p)" :disabled="busy">
            Reverse
          </button></span
        >
      </div></AppPanel
    >
  </div>
</template>
