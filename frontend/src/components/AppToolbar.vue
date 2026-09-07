<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { BarChart3, ChevronDown, FilePlus2, PanelLeftClose, PanelLeftOpen, Printer, ShoppingCart, Store } from 'lucide-vue-next'
import type { CurrencyUnit } from '../utils/currency'
import IconButton from './IconButton.vue'
import SearchField from './SearchField.vue'
import SelectField from './SelectField.vue'

const searchField = ref<{ focus: () => void } | null>(null)

function focusSearch(event: KeyboardEvent) {
  if (event.ctrlKey && event.key.toLowerCase() === 'f') {
    event.preventDefault()
    searchField.value?.focus()
  }
}

onMounted(() => window.addEventListener('keydown', focusSearch))
onBeforeUnmount(() => window.removeEventListener('keydown', focusSearch))

defineProps<{
  collapsed: boolean
  searchQuery: string
  currencyUnit: CurrencyUnit
}>()

defineEmits<{
  'toggle-sidebar': []
  'update:search-query': [value: string]
  'update:currency-unit': [value: CurrencyUnit]
  'new-order': []
  navigate: [view: string]
}>()
</script>

<template>
  <header class="navbar min-h-16 gap-3 border-b border-base-300 bg-base-100 px-4 py-3 lg:px-6">
    <IconButton data-drawer-toggle="true" class="btn-square" :label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      @click.stop="$emit('toggle-sidebar')">
      <PanelLeftOpen v-if="collapsed" :size="18" :stroke-width="1.8" aria-hidden="true" />
      <PanelLeftClose v-else :size="18" :stroke-width="1.8" aria-hidden="true" />
    </IconButton>
    <SearchField
      ref="searchField"
      class="flex-1"
      :model-value="searchQuery"
      placeholder="Search orders, customers, materials..."
      shortcut="Ctrl F"
      aria-label="Global search"
      @update:model-value="$emit('update:search-query', $event)"
    />
    <button class="btn btn-ghost hidden" type="button" aria-label="Current shop: Central shop">
      <Store class="text-primary" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <span><strong>Central shop</strong> · Tehran</span>
      <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <SelectField
      class="w-24"
      :model-value="currencyUnit"
      :options="[{ label: 'Toman', value: 'Toman' }, { label: 'Rial', value: 'Rial' }]"
      aria-label="Display currency"
      @update:model-value="$emit('update:currency-unit', $event as CurrencyUnit)"
    />
    <nav class="hidden items-center gap-1 lg:flex" aria-label="Quick actions">
      <button class="btn btn-ghost btn-square" type="button" aria-label="Production" title="Production" @click="$emit('navigate', 'Production')">
        <Printer :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button class="btn btn-ghost btn-square" type="button" aria-label="Purchases" title="Purchases" @click="$emit('navigate', 'Purchases')">
        <ShoppingCart :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button class="btn btn-ghost btn-square" type="button" aria-label="Reports" title="Reports" @click="$emit('navigate', 'Reports')">
        <BarChart3 :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <button class="btn btn-primary btn-square" type="button" aria-label="New order" title="New order" @click="$emit('new-order')">
        <FilePlus2 :size="16" :stroke-width="1.8" aria-hidden="true" />
      </button>
    </nav>
    <div class="avatar placeholder hidden">
      <div class="w-8 rounded-full bg-primary text-primary-content" aria-hidden="true"><span>RR</span></div>
    </div>
    <div class="hidden text-xs"><strong>Reza Rasoulzadeh</strong><span class="block text-base-content/60">Owner</span>
    </div>
  </header>
</template>
