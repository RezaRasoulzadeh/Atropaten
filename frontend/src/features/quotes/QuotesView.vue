<script setup lang="ts">
import RegisterList from '../../components/ui/RegisterList.vue';
import RegisterRow from '../../components/ui/RegisterRow.vue';
import SearchFilterBar from '../../components/ui/SearchFilterBar.vue';
import LoadingState from '../../components/ui/LoadingState.vue';
import EmptyState from '../../components/ui/EmptyState.vue';
import InlineAlert from '../../components/ui/InlineAlert.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import DataTableCell from '../../components/ui/DataTableCell.vue';
import DataTable from '../../components/ui/DataTable.vue';
import { computed, ref } from 'vue';
import { Plus, SearchX } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import SearchField from '../../components/ui/SearchField.vue';
import SelectField from '../../components/ui/SelectField.vue';
import type { QuoteRecord } from '../../api/quotes';
import { formatMoney, type CurrencyUnit } from '../../utils/currency';
import { formatDateTime } from '../../utils/date';
const props = defineProps<{
  quotes: QuoteRecord[];
  currencyUnit: CurrencyUnit;
  loading?: boolean;
  error?: string;
}>();
const emit = defineEmits<{ openQuote: [id: string]; newQuote: [] }>();
const query = ref('');
const status = ref('All');
const filtered = computed(() =>
  props.quotes.filter(
    (q) =>
      (!query.value.trim() ||
        [q.quoteNumber, q.customerName, ...q.items.map((i) => i.serviceName)].some((v) =>
          v.toLowerCase().includes(query.value.trim().toLowerCase()),
        )) &&
      (status.value === 'All' || q.status === status.value),
  ),
);
function tone(v: string) {
  return v === 'Accepted' || v === 'Converted'
    ? 'green'
    : v === 'Sent'
      ? 'blue'
      : v === 'Rejected'
        ? 'red'
        : v === 'Expired'
          ? 'amber'
          : 'slate';
}
</script>
<template>
  <div class="space-y-4">
    <WorkspaceStickyStack>
      <WorkspaceHeader
        eyebrow="Sales / offers"
        title="Quotes"
        description="Prepare offers and reopen saved pricing snapshots."
        ><button class="btn btn-primary" @click="emit('newQuote')">
          <Plus :size="16" />New quote
        </button></WorkspaceHeader
      >
      <SearchFilterBar
        ><template #search
          ><SearchField
            v-model="query"
            label="Search quotes"
            placeholder="Quote, customer, or service" /></template
        ><template #filters
          ><SelectField
            v-model="status"
            label="Status"
            :options="
              ['All', 'Draft', 'Sent', 'Accepted', 'Rejected', 'Expired', 'Converted'].map(
                (value) => ({ label: value, value }),
              )
            " /></template
        ><template #count>{{ filtered.length }} quotes</template></SearchFilterBar
      >
    </WorkspaceStickyStack>
    <RegisterList
      title="Quote register"
      subtitle="Open an offer to review items, expiry and commercial status."
      :count="filtered.length"
    >
      <LoadingState v-if="loading" label="Loading quotes…" />
      <InlineAlert v-else-if="error" :message="error" />
      <EmptyState
        v-else-if="!filtered.length"
        title="No quotes in this view"
        description="Adjust the filters or create a quote."
      />
      <template v-else
        ><RegisterRow
          v-for="quote in filtered"
          :key="quote.id"
          @activate="emit('openQuote', quote.id)"
        >
          <template #identity
            ><div class="flex min-w-0 flex-wrap items-baseline gap-x-3 gap-y-1">
              <strong class="text-sm">{{ quote.quoteNumber }}</strong
              ><span class="text-sm font-semibold">{{
                quote.customerName || 'Walk-in customer'
              }}</span>
            </div></template
          >
          <template #meta
            ><p class="mt-1 text-xs leading-5 text-base-content/60">
              {{ quote.items.map((i) => i.serviceName).join(' · ') || 'No configured items' }}
            </p>
            <p class="text-xs leading-5 text-base-content/50">
              {{ quote.items.length }} items ·
              {{
                quote.expiryDate ? 'Expires ' + formatDateTime(quote.expiryDate) : 'No expiry date'
              }}
            </p></template
          >
          <template #status
            ><div class="space-y-1 text-end">
              <strong class="block whitespace-nowrap text-sm tabular-nums">{{
                formatMoney(quote.totalRial, currencyUnit)
              }}</strong
              ><StatusBadge :label="quote.status" :tone="tone(quote.status)" /></div
          ></template> </RegisterRow
      ></template>
    </RegisterList>
  </div>
</template>
