<script setup lang="ts">
import { Bell, ChevronDown, Menu, Search, Store } from 'lucide-vue-next'
import type { CurrencyUnit } from '../utils/currency'
import IconButton from './IconButton.vue'
import AppInput from './AppInput.vue'
import AppSelect from './AppSelect.vue'

defineProps<{
  collapsed: boolean
  searchQuery: string
  currencyUnit: CurrencyUnit
}>()

defineEmits<{
  'toggle-sidebar': []
  'update:search-query': [value: string]
  'update:currency-unit': [value: CurrencyUnit]
  notifications: []
}>()
</script>

<template>
  <header class="navbar min-h-16 border-b border-base-300 bg-base-100">
    <IconButton class="btn-square" :label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="$emit('toggle-sidebar')">
      <Menu :size="18" :stroke-width="1.8" aria-hidden="true" />
    </IconButton>
    <label class="relative min-w-0 flex-1">
      <span class="sr-only">Search</span>
      <Search class="pointer-events-none absolute start-3 top-1/2 -translate-y-1/2 text-base-content/50" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <AppInput :model-value="searchQuery" type="search" placeholder="Search orders, customers, materials..." autocomplete="off" @update:model-value="$emit('update:search-query', $event)" />
      <span class="pointer-events-none absolute end-3 top-1/2 -translate-y-1/2 text-xs text-base-content/50">Ctrl K</span>
    </label>
    <span></span>
    <button class="btn btn-ghost hidden" type="button" aria-label="Current shop: Central shop">
      <Store class="text-primary" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <span><strong>Central shop</strong> · Tehran</span>
      <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <label class="relative w-24">
      <span class="sr-only">Display currency</span>
      <AppSelect :model-value="currencyUnit" aria-label="Display currency" @update:model-value="$emit('update:currency-unit', $event as CurrencyUnit)">
        <option value="Toman">Toman</option>
        <option value="Rial">Rial</option>
      </AppSelect>
      <ChevronDown class="pointer-events-none absolute end-2 top-1/2 -translate-y-1/2" :size="13" :stroke-width="1.8" aria-hidden="true" />
    </label>
    <IconButton class="relative" label="Notifications" @click="$emit('notifications')">
      <Bell :size="17" :stroke-width="1.8" aria-hidden="true" />
      <span class="badge badge-xs badge-error absolute end-1 top-1" aria-hidden="true"></span>
    </IconButton>
    <div class="avatar placeholder hidden">
      <div class="w-8 rounded-full bg-primary text-primary-content" aria-hidden="true"><span>RR</span></div>
    </div>
    <div class="hidden text-xs"><strong>Reza Rasoulzadeh</strong><span class="block text-base-content/60">Owner</span>
    </div>
  </header>
</template>
