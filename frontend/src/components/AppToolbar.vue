<script setup lang="ts">
import { Bell, ChevronDown, Menu, Search, Store } from 'lucide-vue-next'
import type { CurrencyUnit } from '../utils/currency'

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
    <button class="sidebar-toggle" type="button" :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="$emit('toggle-sidebar')">
      <Menu :size="18" :stroke-width="1.8" aria-hidden="true" />
    </button>
    <label class="topbar-search">
      <span class="sr-only">Search</span>
      <Search class="search-icon" :size="16" :stroke-width="1.8" aria-hidden="true" />
      <input
        :value="searchQuery"
        type="search"
        placeholder="Search orders, customers, materials..."
        autocomplete="off"
        @input="$emit('update:search-query', ($event.target as HTMLInputElement).value)"
      />
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
      <select :value="currencyUnit" aria-label="Display currency" @change="$emit('update:currency-unit', ($event.target as HTMLSelectElement).value as CurrencyUnit)">
        <option value="Toman">Toman</option>
        <option value="Rial">Rial</option>
      </select>
      <ChevronDown class="toolbar-currency-chevron" :size="13" :stroke-width="1.8" aria-hidden="true" />
    </label>
    <button class="icon-button notification-button" type="button" aria-label="Notifications" @click="$emit('notifications')">
      <Bell :size="17" :stroke-width="1.8" aria-hidden="true" />
      <span class="notification-dot" aria-hidden="true"></span>
    </button>
    <div class="user-profile">
      <span class="avatar" aria-hidden="true">RR</span>
      <span class="user-copy"><strong>Reza Rasoulzadeh</strong><span>Owner</span></span>
    </div>
  </header>
</template>
