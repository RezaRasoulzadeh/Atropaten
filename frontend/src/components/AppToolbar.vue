<script setup lang="ts">
import { BarChart3, ChevronDown, FilePlus2, PanelLeftClose, PanelLeftOpen, Printer, Search, ShoppingCart, Store } from 'lucide-vue-next'
import type { CurrencyUnit } from '../utils/currency'
import IconButton from './IconButton.vue'

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
    <IconButton class="btn-square" :label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
      @click="$emit('toggle-sidebar')">
      <PanelLeftOpen v-if="collapsed" :size="18" :stroke-width="1.8" aria-hidden="true" />
      <PanelLeftClose v-else :size="18" :stroke-width="1.8" aria-hidden="true" />
    </IconButton>
    <label class="relative min-w-0 flex-1">
      <span class="sr-only">Search</span>
      <Search class="pointer-events-none absolute inset-s-3 top-1/2 -translate-y-1/2 text-base-content/50" :size="16"
        :stroke-width="1.8" aria-hidden="true" />
      <input class="input input-bordered w-full min-w-0" :value="searchQuery" type="search"
        placeholder="Search orders, customers, materials..." autocomplete="off"
        @input="$emit('update:search-query', ($event.target as HTMLInputElement).value)" />
      <span class="pointer-events-none absolute inset-e-3 top-1/2 -translate-y-1/2 text-xs text-base-content/50">Ctrl
        K</span>
    </label>
    <button class="btn btn-ghost hidden" type="button" aria-label="Current shop: Central shop">
      <Store class="text-primary" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <span><strong>Central shop</strong> · Tehran</span>
      <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <div class="dropdown dropdown-end">
      <button class="btn btn-outline w-24 justify-between" type="button" tabindex="0"
        aria-haspopup="listbox" aria-label="Display currency">
        <span>{{ currencyUnit }}</span>
        <ChevronDown :size="13" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <ul class="menu dropdown-content z-10 mt-1 w-24 rounded-box border border-base-300 bg-base-100 p-1 shadow-none"
        role="listbox" aria-label="Display currency options" tabindex="0">
        <li><button type="button" role="option" :aria-selected="currencyUnit === 'Toman'"
            @click="$emit('update:currency-unit', 'Toman')">Toman</button></li>
        <li><button type="button" role="option" :aria-selected="currencyUnit === 'Rial'"
            @click="$emit('update:currency-unit', 'Rial')">Rial</button></li>
      </ul>
    </div>
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
