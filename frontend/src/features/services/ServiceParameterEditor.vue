<script setup lang="ts">
import FormGrid from '../../components/ui/FormGrid.vue'
import FormField from '../../components/ui/FormField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import { ref } from 'vue'
import {ChevronUp,ChevronDown,ChevronRight,Trash2,Plus} from 'lucide-vue-next'
import type {ParameterForm} from './types'
import type {MaterialRecord} from '../../api/materials'
const props=defineProps<{parameter:ParameterForm;index:number;count:number;materials:MaterialRecord[];showErrors?:boolean}>()
const emit=defineEmits<{move:[direction:-1|1];remove:[];normalize:[];addOption:[];removeOption:[index:number]}>()
const expanded=ref(props.index === 0)
function syncExpanded(event: Event) {
  expanded.value = (event.target as HTMLDetailsElement).open
}
</script>
<template><details class="group min-w-0 border-t border-base-300" :open="expanded" @toggle="syncExpanded">
              <summary class="flex min-w-0 cursor-pointer list-none items-center justify-between gap-3 px-4 py-3 [&::-webkit-details-marker]:hidden">
                <div class="flex min-w-0 items-center gap-3">
                  <span class="grid size-7 shrink-0 place-items-center rounded-full bg-primary/15 text-xs font-semibold text-primary">{{ String(index + 1).padStart(2, '0') }}</span>
                  <div class="min-w-0">
                    <strong class="block truncate">{{ parameter.label || 'Untitled parameter' }}</strong>
                    <small class="block truncate text-xs text-base-content/60"><code>{{ parameter.key || 'key not set' }}</code> · {{ parameter.type }}</small>
                  </div>
                </div>
                <div class="flex flex-wrap items-center gap-2">
                  <button
                    class="btn btn-ghost btn-sm"
                    type="button"
                    :disabled="index === 0"
                    :aria-label="`Move ${parameter.label || 'parameter'} up`"
                    @click.stop.prevent="emit('move',-1)"
                  >
                    <ChevronUp :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-ghost btn-sm"
                    type="button"
                    :disabled="index === count - 1"
                    :aria-label="`Move ${parameter.label || 'parameter'} down`"
                    @click.stop.prevent="emit('move',1)"
                  >
                    <ChevronDown :size="14" :stroke-width="1.8" aria-hidden="true" /></button
                  ><button
                    class="btn btn-outline btn-error btn-sm"
                    type="button"
                    :aria-label="`Remove ${parameter.label || 'parameter'}`"
                    @click.stop.prevent="emit('remove')"
                  >
                    <Trash2 :size="14" :stroke-width="1.8" aria-hidden="true" />
                  </button><ChevronRight class="shrink-0 transition-transform group-open:rotate-90" :size="15" :stroke-width="1.8" aria-hidden="true" />
                </div>
              </summary>
              <div class="min-w-0 space-y-3 border-t border-base-300 px-4 py-4">
              <FormGrid
                ><FormField class="gap-1"
                  ><span>Key</span
                  ><AppInput
                    class="input w-full min-w-0"
                    :class="{ 'input-error': props.showErrors && !parameter.key.trim() }"
                    v-model="parameter.key"
                    type="text"
                    required
                    placeholder="paper_size"
                    autocomplete="off" /></FormField
                ><FormField class="gap-1"
                  ><span>Label</span
                  ><AppInput
                    class="input w-full min-w-0"
                    :class="{ 'input-error': props.showErrors && !parameter.label.trim() }"
                    v-model="parameter.label"
                    type="text"
                    required
                    placeholder="Paper size"
                    autocomplete="off" /></FormField
              ></FormGrid>
              <FormGrid
                ><SelectField
                  v-model="parameter.type"
                  :invalid="props.showErrors && !parameter.type"
                  label="Type"
                  :options="[
                    { label: 'Integer', value: 'integer' },
                    { label: 'Decimal', value: 'decimal' },
                    { label: 'Boolean', value: 'boolean' },
                    { label: 'Choice', value: 'choice' },
                    { label: 'Material reference', value: 'material-reference' },
                  ]"
                  @update:model-value="emit('normalize')"
                /><FormField class="gap-1"
                  ><span>Required</span
                  ><span
                    ><input class="checkbox" v-model="parameter.required" type="checkbox" />Required
                    input</span
                  ></FormField
                ></FormGrid
              >
              <FormGrid v-if="parameter.type === 'integer' || parameter.type === 'decimal'"
                ><FormField class="gap-1"
                  ><span>Minimum</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.minValue"
                    type="text"
                    inputmode="decimal"
                    placeholder="Optional" /></FormField
                ><FormField class="gap-1"
                  ><span>Maximum</span
                  ><AppInput
                    class="input w-full min-w-0"
                    v-model="parameter.maxValue"
                    type="text"
                    inputmode="decimal"
                    placeholder="Optional" /></FormField
              ></FormGrid>
              <FormField
                class="gap-1"
                v-if="parameter.type === 'integer' || parameter.type === 'decimal'"
                ><span>Default value</span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="parameter.defaultValue"
                  type="text"
                  inputmode="decimal"
                  placeholder="Optional"
              /></FormField>
              <SelectField
                v-else-if="parameter.type === 'boolean'"
                v-model="parameter.defaultValue"
                label="Default value"
                :options="[
                  { label: 'Not set', value: '' },
                  { label: 'True', value: 'true' },
                  { label: 'False', value: 'false' },
                ]"
              />
              <SelectField
                v-else-if="parameter.type === 'choice'"
                v-model="parameter.defaultValue"
                label="Default choice"
                :options="[
                  { label: 'Not set', value: '' },
                  ...parameter.options.map((option) => ({ label: option, value: option })),
                ]"
              />
              <div v-else class="min-w-0 space-y-3">
                <SelectField
                  v-model="parameter.defaultValue"
                  label="Default material"
                  :options="[
                    { label: 'No default material', value: '' },
                    ...materials.map((material) => ({
                      label: `${material.name}${material.sku ? ` · ${material.sku}` : ''}`,
                      value: material.id,
                    })),
                  ]"
                /><small class="block text-xs leading-5 text-base-content/60"
                  >Use this parameter when a material cost should follow the operator's selected
                  paper or stock item.</small
                >
              </div>
              <div v-if="parameter.type === 'choice'" class="min-w-0 space-y-3">
                <div class="min-w-0 space-y-3">
                  <span>Choice options</span
                  ><button class="btn btn-ghost" type="button" @click="emit('addOption')">
                    <Plus :size="14" :stroke-width="1.8" aria-hidden="true" />Add option
                  </button>
                </div>
                <div
                  v-for="(option, optionIndex) in parameter.options"
                  :key="`${parameter.id}-${optionIndex}`"
                  class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
                >
                  <AppInput
                    class="input w-full min-w-0"
                    :class="{ 'input-error': props.showErrors && !parameter.options[optionIndex].trim() }"
                    v-model="parameter.options[optionIndex]"
                    type="text"
                    required
                    :aria-label="`Choice option ${optionIndex + 1}`"
                    placeholder="A4"
                    @input="
                      parameter.defaultValue =
                        parameter.defaultValue === option
                          ? parameter.options[optionIndex]
                          : parameter.defaultValue
                    "
                  /><button
                    class="btn btn-outline btn-error"
                    type="button"
                    :aria-label="`Remove choice option ${optionIndex + 1}`"
                    @click="emit('removeOption',optionIndex)"
                  >
                    <Trash2 :size="13" :stroke-width="1.8" aria-hidden="true" />
                  </button>
                </div>
                <p v-if="!parameter.options.length">Add at least one option before saving.</p>
              </div>
              <FormField class="gap-1"
                ><span>Unit / suffix <em>optional</em></span
                ><AppInput
                  class="input w-full min-w-0"
                  v-model="parameter.unit"
                  type="text"
                  placeholder="sheets, hours, or mm"
              /></FormField>
              </div>
            </details></template>
