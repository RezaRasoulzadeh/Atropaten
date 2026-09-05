<script setup lang="ts">
defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <article class="panel">
    <header class="panel-heading">
      <div>
        <h2 class="panel-title">{{ title }}</h2>
        <p v-if="subtitle" class="panel-subtitle">{{ subtitle }}</p>
      </div>
      <div v-if="$slots.action" class="panel-action">
        <slot name="action" />
      </div>
    </header>
    <div class="panel-body">
      <slot />
    </div>
  </article>
</template>
