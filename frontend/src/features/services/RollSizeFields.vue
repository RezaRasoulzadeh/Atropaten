<script setup lang="ts">
import { ArrowLeftRight } from 'lucide-vue-next'
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
const props = defineProps<{ width: string; height: string; maxWidth?: string; materialName?: string; widthInvalid?: boolean }>()
const emit = defineEmits<{ width: [value: string]; height: [value: string] }>()

function swapDimensions() {
  emit('width', props.height)
  emit('height', props.width)
}
</script>
<template>
  <div class="grid grid-cols-1 items-center gap-3 sm:grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] sm:gap-4" :aria-label='$t("Roll dimensions")'>
    <FormField class="gap-1">
      <span>{{ $t("Custom width (mm)") }} <em class="text-error">*</em></span>
      <AppInput :model-value="width" :class="{ 'input-error': widthInvalid }" type="number" min="0.001" :max="maxWidth || undefined" step="any" inputmode="decimal" :aria-label='$t("Custom width (mm)")' :placeholder='$t("e.g. 1000")' @update:model-value="emit('width', $event)" />
      <small v-if="widthInvalid" class="text-xs text-error">{{ $t("Enter a positive width no greater than the selected material width (") }}{{ $ui(maxWidth || 'select a roll') }} {{ $t("mm).") }}</small>
      <small v-else class="text-xs text-base-content/60">{{ $ui(materialName ? `${materialName} · maximum ${maxWidth || '—'} mm before edge margins` : 'Select a material with a configured roll width.') }}</small>
    </FormField>
    <button class="btn btn-ghost btn-square btn-sm justify-self-center text-primary" type="button" :aria-label='$t("Swap width and height")' :title='$t("Swap width and height")' @click="swapDimensions">
      <ArrowLeftRight :size="17" aria-hidden="true" />
    </button>
    <FormField class="gap-1">
      <span>{{ $t("Custom height / length (mm)") }} <em class="text-error">*</em></span>
      <AppInput :model-value="height" type="number" min="0.001" step="any" inputmode="decimal" :aria-label='$t("Custom height / length (mm)")' :placeholder='$t("e.g. 3000")' @update:model-value="emit('height', $event)" />
      <small class="text-xs text-base-content/60">{{ $t("Enter the required length along the roll. 1000 mm = 1 m.") }}</small>
    </FormField>
  </div>
</template>
