<script setup lang="ts">
import AppInput from '../../components/ui/AppInput.vue'
import FormField from '../../components/ui/FormField.vue'
defineProps<{ width: string; height: string; materialName?: string }>()
const emit = defineEmits<{ height: [value: string] }>()
</script>
<template>
  <div class="grid gap-4 sm:grid-cols-2" aria-label="Roll dimensions">
    <FormField class="gap-1">
      <span>Width from material (mm)</span>
      <AppInput :model-value="width" readonly aria-label="Width from material (mm)" placeholder="Select a roll material" />
      <small class="text-xs text-base-content/60">{{ materialName ? `${materialName} · usable width after edge margins` : 'Select a material with a configured roll width.' }}</small>
    </FormField>
    <FormField class="gap-1">
      <span>Custom height / length (mm) <em class="text-error">*</em></span>
      <AppInput :model-value="height" type="number" min="0.001" step="any" inputmode="decimal" aria-label="Custom height / length (mm)" placeholder="e.g. 3000" @update:model-value="emit('height', $event)" />
      <small class="text-xs text-base-content/60">Enter the required length along the roll. 1000 mm = 1 m.</small>
    </FormField>
  </div>
</template>
