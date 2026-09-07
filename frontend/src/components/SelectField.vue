<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
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
const root = ref<HTMLElement | null>(null)
const isOpen = ref(false)
const selected = computed(() => props.options.find((option) => option.value === props.modelValue) ?? props.options[0])

function toggle() {
  if (!props.options.length) return
  isOpen.value = !isOpen.value
}

function choose(value: string) {
  emit('update:modelValue', value)
  isOpen.value = false
  trigger.value?.blur()
}

function onDocumentClick(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) isOpen.value = false
}

function onKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') isOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', onDocumentClick)
  document.addEventListener('keydown', onKeydown)
})

onBeforeUnmount(() => {
  document.removeEventListener('click', onDocumentClick)
  document.removeEventListener('keydown', onKeydown)
})
</script>

<template>
  <div class="form-control min-w-0 gap-1">
    <span v-if="label" class="text-xs text-base-content/60">{{ label }}</span>
    <div ref="root" class="relative w-full">
      <button
        ref="trigger"
        class="select-field-trigger btn btn-outline h-10 min-h-10 w-full justify-between gap-2 px-3 font-normal normal-case"
        type="button"
        tabindex="0"
        :disabled="!options.length"
        :aria-label="ariaLabel || label || 'Select an option'"
        aria-haspopup="listbox"
        :aria-expanded="isOpen"
        @click.stop="toggle"
      >
        <span class="truncate">{{ selected?.label ?? (props.options.length ? props.modelValue : 'No options available') }}</span>
        <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" />
      </button>
      <ul
        v-if="isOpen"
        tabindex="0"
        class="absolute inset-x-0 top-full z-50 mt-1 max-h-60 w-full overflow-y-auto rounded-box border border-base-300 bg-base-100 p-1 shadow-lg"
        role="listbox"
        :aria-label="ariaLabel || label || 'Select an option'"
        @click.stop
      >
        <li v-for="option in options" :key="option.value">
          <button
            type="button"
            role="option"
            :aria-selected="option.value === modelValue"
            class="flex w-full items-center justify-between gap-2 rounded-field px-3 py-2 text-start text-sm hover:bg-base-200"
            @click.stop="choose(option.value)"
          >
            <span>{{ option.label }}</span>
            <Check v-if="option.value === modelValue" :size="14" :stroke-width="2" aria-hidden="true" />
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
