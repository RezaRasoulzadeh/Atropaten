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
  <header class="topbar">
    <IconButton class="sidebar-toggle" :label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="$emit('toggle-sidebar')">
      <Menu :size="18" :stroke-width="1.8" aria-hidden="true" />
    </IconButton>
    <label class="topbar-search">
      <span class="sr-only">Search</span>
      <Search class="search-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <AppInput :model-value="searchQuery" type="search" placeholder="Search orders, customers, materials..." autocomplete="off" @update:model-value="$emit('update:search-query', $event)" />
      <span class="search-shortcut">Ctrl K</span>
    </label>
    <span class="topbar-spacer"></span>
    <button class="shop-selector" type="button" aria-label="Current shop: Central shop">
      <Store class="shop-dot" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <span><strong>Central shop</strong> · Tehran</span>
      <ChevronDown class="shop-selector-chevron" :size="14" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <label class="toolbar-currency">
      <span class="sr-only">Display currency</span>
      <AppSelect :model-value="currencyUnit" aria-label="Display currency" @update:model-value="$emit('update:currency-unit', $event as CurrencyUnit)">
        <option value="Toman">Toman</option>
        <option value="Rial">Rial</option>
      </AppSelect>
      <ChevronDown class="toolbar-currency-chevron" :size="13" :stroke-width="1.8" aria-hidden="true" />
    </label>
    <IconButton class="notification-button" label="Notifications" @click="$emit('notifications')">
      <Bell :size="17" :stroke-width="1.8" aria-hidden="true" />
      <span class="notification-dot" aria-hidden="true"></span>
    </IconButton>
    <div class="user-profile">
      <span class="avatar" aria-hidden="true">RR</span>
      <span class="user-copy"><strong>Reza Rasoulzadeh</strong><span>Owner</span></span>
    </div>
  </header>
</template>
