<script setup lang="ts">
import { computed, ref } from 'vue'
import { Check, ChevronDown } from 'lucide-vue-next'

type SelectOption = {
  label: string
  value: string
}

const props = withDefaults(defineProps<{
  modelValue: string
  options: readonly SelectOption[]
  label?: string
  ariaLabel?: string
}>(), {
  label: '',
  ariaLabel: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const trigger = ref<HTMLButtonElement | null>(null)
const selected = computed(() => props.options.find((option) => option.value === props.modelValue) ?? props.options[0])

function choose(value: string) {
  emit('update:modelValue', value)
  trigger.value?.blur()
}
</script>

<template>
  <div class="form-control min-w-0 gap-1">
    <span v-if="label" class="text-xs text-base-content/60">{{ label }}</span>
    <div class="dropdown dropdown-bottom w-full">
      <button
        ref="trigger"
        class="btn btn-outline h-10 min-h-10 w-full justify-between gap-2 px-3 font-normal normal-case"
        type="button"
        tabindex="0"
        :aria-label="ariaLabel || label || 'Select an option'"
        aria-haspopup="listbox"
      >
        <span class="truncate">{{ selected?.label ?? props.modelValue }}</span>
        <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <ul
        tabindex="0"
        class="menu dropdown-content z-30 mt-1 w-full min-w-full rounded-box border border-base-300 bg-base-100 p-1 shadow-lg"
        role="listbox"
        :aria-label="ariaLabel || label || 'Select an option'"
      >
        <li v-for="option in options" :key="option.value">
          <button
            type="button"
            role="option"
            :aria-selected="option.value === modelValue"
            class="justify-between"
            @click="choose(option.value)"
          >
            <span>{{ option.label }}</span>
            <Check v-if="option.value === modelValue" :size="14" :stroke-width="2" aria-hidden="true" />
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
