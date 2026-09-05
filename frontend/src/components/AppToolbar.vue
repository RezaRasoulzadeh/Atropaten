<script setup lang="ts">
import AppIcon from './AppIcon.vue'

defineProps<{
  collapsed: boolean
  searchQuery: string
}>()

defineEmits<{
  'toggle-sidebar': []
  'update:search-query': [value: string]
  notifications: []
}>()
</script>

<template>
  <header class="topbar">
    <button class="sidebar-toggle" type="button" :aria-label="collapsed ? 'Expand sidebar' : 'Collapse sidebar'" @click="$emit('toggle-sidebar')">
      <AppIcon name="menu" :size="18" />
    </button>
    <label class="topbar-search">
      <span class="sr-only">Search</span>
      <AppIcon class="search-icon" name="search" :size="16" />
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
      <span class="shop-dot" aria-hidden="true"></span>
      <span><strong>Central shop</strong> · Tehran</span>
      <span class="shop-selector-chevron" aria-hidden="true">⌄</span>
    </button>
    <button class="icon-button notification-button" type="button" aria-label="Notifications" @click="$emit('notifications')">
      <AppIcon name="bell" :size="17" />
      <span class="notification-dot" aria-hidden="true"></span>
    </button>
    <div class="user-profile">
      <span class="avatar" aria-hidden="true">RR</span>
      <span class="user-copy"><strong>Reza Rasoulzadeh</strong><span>Owner</span></span>
    </div>
  </header>
</template>
