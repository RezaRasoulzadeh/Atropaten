<script setup lang="ts" generic="T extends string">
import { computed, onBeforeUnmount, onMounted, ref, useId, watch } from 'vue';
defineOptions({ inheritAttrs: false });
const props = defineProps<{
  modelValue: T;
  options: readonly { label: string; value: T; disabled?: boolean }[];
  label?: string;
  ariaLabel?: string;
  disabled?: boolean;
  invalid?: boolean;
}>();
const emit = defineEmits<{ 'update:modelValue': [value: T] }>();
const id = useId();
const root = ref<HTMLElement | null>(null);
const open = ref(false);
const suppressNextToggle = ref(false);
const selectedOption = computed(() => props.options.find((option) => option.value === props.modelValue));
const isDisabled = computed(() => props.disabled || !props.options.length);

watch(
  () => props.modelValue,
  () => {
    open.value = false;
  },
);

function closeOnOutsideClick(event: MouseEvent) {
  if (root.value && !root.value.contains(event.target as Node)) open.value = false;
}

function toggle() {
  if (suppressNextToggle.value) {
    suppressNextToggle.value = false;
    open.value = false;
    return;
  }
  if (!isDisabled.value) open.value = !open.value;
}

function choose(value: T) {
  suppressNextToggle.value = true;
  open.value = false;
  emit('update:modelValue', value);
  queueMicrotask(() => {
    suppressNextToggle.value = false;
  });
}

onMounted(() => document.addEventListener('click', closeOnOutsideClick));
onBeforeUnmount(() => document.removeEventListener('click', closeOnOutsideClick));
</script>
<template>
  <div ref="root" class="flex min-w-0 flex-col gap-1">
    <label
      v-if="label"
      :for="String($attrs.id || id)"
      class="text-xs leading-4"
      :class="invalid ? 'text-error' : 'text-base-content/65'"
      >{{ label }}</label
    >
    <div class="dropdown relative w-full" :class="{ 'dropdown-open': open }">
      <button
        v-bind="$attrs"
        :id="String($attrs.id || id)"
        class="select h-10 min-h-10 w-full min-w-0 text-start text-sm font-normal leading-5"
        :class="{ 'select-error': invalid }"
        type="button"
        :disabled="isDisabled"
        :aria-label="ariaLabel || label || 'Select an option'"
        aria-haspopup="listbox"
        :aria-expanded="open"
        :aria-invalid="invalid || undefined"
        @click="toggle"
        @keydown.esc.prevent="open = false"
        @keydown.down.prevent="open = true"
      >
        {{ selectedOption?.label || (!options.length ? 'No options available' : '') }}
      </button>
      <ul
        v-if="open"
        class="dropdown-content menu z-50 mt-1 max-h-60 w-max min-w-full flex-nowrap overflow-y-auto rounded-box border border-base-300 bg-base-100 p-1 shadow-lg"
        role="listbox"
        :aria-label="ariaLabel || label || 'Select an option'"
      >
        <li v-for="option in options" :key="option.value">
          <button
            type="button"
            role="option"
            :disabled="option.disabled"
            :aria-selected="option.value === modelValue"
            class="min-h-9 text-start text-sm"
            :class="option.value === modelValue ? 'active' : ''"
            @click.stop.prevent="choose(option.value)"
          >
            {{ option.label }}
          </button>
        </li>
      </ul>
    </div>
  </div>
</template>
