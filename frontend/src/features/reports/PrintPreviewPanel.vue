<script setup lang="ts">
import { ref } from 'vue'
import { Printer, FileText } from 'lucide-vue-next'
import AppPanel from '../../components/layout/AppPanel.vue'
import FormField from '../../components/ui/FormField.vue'
import AppInput from '../../components/ui/AppInput.vue'
import SelectField from '../../components/ui/SelectField.vue'
import InlineAlert from '../../components/ui/InlineAlert.vue'
import PrintDocument from './PrintDocument.vue'
import { reportsApi, type PrintDocumentRecord } from '../../api/reports'
import { useWorkspaceActions } from '../../composables/useWorkspaceActions'
import { normalizeError, useToast } from '../../ui/feedback'
import type { CurrencyUnit } from '../../utils/currency'
const props=defineProps<{ currencyUnit: CurrencyUnit; start: string | null; end: string | null }>()
const kind=ref('invoice'),recordId=ref(''),partyId=ref(''),error=ref('')
const document=ref<PrintDocumentRecord|null>(null)
const {busy,runAction}=useWorkspaceActions()
const toast=useToast()
async function loadPreview() {
 await runAction(async()=>{
  error.value=''
  const statement=kind.value.includes('statement')
  if(!(statement ? partyId.value : recordId.value).trim()) { error.value=statement?'Enter a party ID.':'Enter a source record ID.'; toast.warning(error.value); return }
  try { document.value=await reportsApi.print(kind.value,recordId.value,props.start||'',props.end||'',partyId.value);toast.success('Print preview loaded.') }
  catch(e) { error.value=normalizeError(e).message;throw e }
 })
}
function printDocument() { window.print() }
</script>
<template>
<AppPanel title="Print preview" subtitle="Load a saved invoice, receipt or statement.">
<form class="grid min-w-0 items-end gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]" @submit.prevent="loadPreview">
<SelectField v-model="kind" label="Document" :options="[{label:'Invoice',value:'invoice'},{label:'Payment receipt',value:'payment_receipt'},{label:'Customer statement',value:'customer_statement'},{label:'Supplier statement',value:'supplier_statement'}]" />
<FormField v-if="kind.includes('statement')" label="Party ID"><AppInput v-model="partyId" required placeholder="Customer or supplier ID" /></FormField><FormField v-else label="Record ID"><AppInput v-model="recordId" required placeholder="Invoice or payment ID" /></FormField>
<button class="btn btn-primary" :disabled="busy"><FileText :size="15" />Load preview</button>
</form>
<InlineAlert v-if="error" :message="error" />
<template v-if="document"><div class="flex justify-end"><button class="btn btn-outline" @click="printDocument"><Printer :size="15" />Print</button></div><PrintDocument :document="document" :currency-unit="currencyUnit" /><Teleport to="body"><div class="print-output"><PrintDocument :document="document" :currency-unit="currencyUnit" /></div></Teleport></template>
</AppPanel>
</template>
