<script setup lang="ts">
import { ref } from 'vue'
import { Search } from 'lucide-vue-next'

withDefaults(defineProps<{
  modelValue: string
  label?: string
  placeholder?: string
  shortcut?: string
  ariaLabel?: string
}>(), {
  label: '',
  placeholder: 'Search',
  shortcut: '',
  ariaLabel: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const input = ref<HTMLInputElement | null>(null)

function focus() {
  input.value?.focus()
}

defineExpose({ focus })
</script>

<template>
  <label class="form-control min-w-0 gap-1">
    <span v-if="label" class="text-xs text-base-content/60">{{ label }}</span>
    <span class="relative block min-w-0">
      <Search
        class="pointer-events-none absolute inset-s-3 top-1/2 z-10 -translate-y-1/2 text-base-content/70"
        :size="15"
        :stroke-width="1.8"
        aria-hidden="true"
      />
      <input
        ref="input"
        class="input input-bordered w-full min-w-0 ps-9"
        :class="shortcut ? 'pe-16' : ''"
        :value="modelValue"
        type="search"
        :placeholder="placeholder"
        :aria-label="ariaLabel || label || undefined"
        autocomplete="off"
        @input="emit('update:modelValue', ($event.target as HTMLInputElement).value)"
      />
      <span
        v-if="shortcut"
        class="pointer-events-none absolute inset-e-3 top-1/2 z-10 -translate-y-1/2 text-xs text-base-content/60"
      >{{ shortcut }}</span>
    </span>
  </label>
</template>
