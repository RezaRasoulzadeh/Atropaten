<script setup lang="ts">
import { Printer } from 'lucide-vue-next'
import StatusBadge from '../../components/ui/StatusBadge.vue'
import type { ServiceForm } from './types'

defineProps<{
  form: ServiceForm
  active: boolean
}>()
</script>

<template>
  <div
    class="relative flex min-h-36 overflow-hidden rounded-box bg-base-300 bg-cover bg-center"
    :style="form.imagePath ? { backgroundImage: `url('${form.imagePath}')` } : undefined"
  >
    <div class="absolute inset-0 bg-gradient-to-t from-black/90 via-black/50 to-black/10" aria-hidden="true"></div>
    <div class="relative z-10 mt-auto flex min-w-0 items-end gap-3 p-3">
      <div v-if="!form.imagePath" class="grid size-10 shrink-0 place-items-center rounded-box bg-black/25 text-white/75">
        <Printer :size="21" aria-hidden="true" />
      </div>
      <div class="min-w-0 text-white">
        <strong class="block truncate text-base font-semibold">{{ form.name || 'Your service name' }}</strong>
        <span class="mt-1 block truncate text-sm text-white/70">
          {{ form.code || 'SVC-001' }}<span v-if="form.category"> · {{ form.category }}</span>
        </span>
        <StatusBadge class="mt-2" :label="active ? 'Active' : 'Archived'" :tone="active ? 'green' : 'slate'" />
      </div>
    </div>
  </div>
</template>
