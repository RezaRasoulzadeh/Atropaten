<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import AppInput from '../../components/ui/AppInput.vue';
import { onMounted, ref } from 'vue';
import { FilePlus2, Plus, Trash2 } from 'lucide-vue-next';
import { metadataApi, type AttachmentRecord, type ProofRecord } from '../../api/metadata';
import { formatDateTime } from '../../utils/date';
import SelectField from '../../components/ui/SelectField.vue';
const props = withDefaults(
  defineProps<{ ownerType: 'quote' | 'order'; ownerId: string; protectedContext?: boolean }>(),
  { protectedContext: false },
);
const emit = defineEmits<{ notify: [message: string] }>();
const attachments = ref<AttachmentRecord[]>([]);
const proofs = ref<ProofRecord[]>([]);
const loading = ref(true);
const fileName = ref('');
const path = ref('');
const category = ref('artwork');
const notes = ref('');
const proofStatus = ref('Draft');
const proofVersion = ref('v1');
const proofNote = ref('');
const selectedAttachment = ref('');
async function load() {
  loading.value = true;
  try {
    [attachments.value, proofs.value] = await Promise.all([
      metadataApi.attachments(props.ownerType, props.ownerId),
      metadataApi.proofs(props.ownerType, props.ownerId),
    ]);
  } catch (e) {
    emit('notify', String(e));
  } finally {
    loading.value = false;
  }
}
async function addFile() {
return runAction(async () => {
  if (!fileName.value.trim() || !path.value.trim()) {
    emit('notify', 'Display name and path are required');
    return;
  }
  try {
    await metadataApi.addAttachment({
      ownerType: props.ownerType,
      ownerId: props.ownerId,
      fileName: fileName.value,
      path: path.value,
      mimeType: '',
      sizeBytes: null,
      checksum: '',
      category: category.value,
      notes: notes.value,
    });
    fileName.value = path.value = notes.value = '';
    await load();
    emit('notify', 'Attachment metadata saved');
  } catch (e) {
reportError(e);
  }

});
}
async function removeFile(id: string) {
return runAction(async () => {
  if (props.protectedContext) {
    emit('notify', 'Attachments on a confirmed document are protected');
    return;
  }
  try {
    await metadataApi.removeAttachment(id);
    await load();
    emit('notify', 'Attachment metadata removed');
  } catch (e) {
reportError(e);
  }

});
}
async function addProof() {
return runAction(async () => {
  try {
    await metadataApi.createProof({
      ownerType: props.ownerType,
      ownerId: props.ownerId,
      attachmentId: selectedAttachment.value,
      status: proofStatus.value,
      versionLabel: proofVersion.value,
      approverNote: proofNote.value,
      internalNote: '',
    });
    proofNote.value = '';
    await load();
    emit('notify', 'Proof version recorded');
  } catch (e) {
reportError(e);
  }

});
}
onMounted(load);
</script>
<template>
  <div class="min-w-0 space-y-3">
    <section class="min-w-0 space-y-4">
      <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
        <div class="min-w-0 space-y-3">
          <h2 class="text-base font-semibold">Files & attachments</h2>
          <p>Metadata and path references only; file bytes stay outside SQLite.</p>
        </div>
        <FilePlus2 :size="17" />
      </header>
      <div class="min-w-0 space-y-3">
        <FormField label="Display name"><AppInput
          class="input w-full min-w-0"
          v-model="fileName"
          placeholder="Display name"
        /></FormField><FormField label="Relative or external path"><AppInput
          class="input w-full min-w-0"
          v-model="path"
          placeholder="Relative or external path"
        /></FormField><SelectField
          v-model="category"
          label="Category"
          :options="[
            { label: 'Artwork', value: 'artwork' },
            { label: 'Proof', value: 'proof' },
            { label: 'Reference', value: 'reference' },
            { label: 'Other', value: 'other' },
          ]"
        /><FormField label="Notes"><AppInput
          class="input w-full min-w-0"
          v-model="notes"
          placeholder="Notes"
        /></FormField><button class="btn btn-ghost" @click="addFile" :disabled="busy"><Plus :size="14" />Add metadata</button>
      </div>
      <LoadingState v-if="loading" label="Loading records…" />
      <div v-else-if="!attachments.length">No attachment metadata yet.</div>
      <div
        v-for="file in attachments"
        :key="file.id"
        class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
      >
        <div class="min-w-0 space-y-1">
          <strong>{{ file.fileName }}</strong
          ><span>{{ file.category }} · {{ file.path }}</span
          ><small v-if="file.notes" class="block text-xs leading-5 text-base-content/60">{{
            file.notes
          }}</small>
        </div>
        <button
          class="btn btn-ghost"
          aria-label="Remove attachment metadata"
          @click="removeFile(file.id)"
         :disabled="busy">
          <Trash2 :size="14" />
        </button>
      </div>
    </section>
    <section class="min-w-0 space-y-4">
      <header class="flex min-w-0 flex-wrap items-start justify-between gap-3">
        <div class="min-w-0 space-y-3">
          <h2 class="text-base font-semibold">Proof history</h2>
          <p>Every status entry is a preserved version of the proof record.</p>
        </div>
      </header>
      <div class="min-w-0 space-y-3">
        <FormField label="Version / label"><AppInput
          class="input w-full min-w-0"
          v-model="proofVersion"
          placeholder="Version / label"
        /></FormField><SelectField
          v-model="selectedAttachment"
          label="Related attachment"
          :options="[
            { label: 'No related attachment', value: '' },
            ...attachments.map((file) => ({ label: file.fileName, value: file.id })),
          ]"
        /><SelectField
          v-model="proofStatus"
          label="Status"
          :options="
            ['Draft', 'Ready', 'Waiting Customer Approval', 'Approved', 'Rejected'].map(
              (value) => ({ label: value, value }),
            )
          "
        /><FormField label="Approval / rejection note"><AppInput
          class="input w-full min-w-0"
          v-model="proofNote"
          placeholder="Approval / rejection note"
        /></FormField><button class="btn btn-ghost" @click="addProof" :disabled="busy"><Plus :size="14" />Record version</button>
      </div>
      <div v-if="!proofs.length">No proof versions recorded yet.</div>
      <div
        v-for="proof in proofs"
        :key="proof.id"
        class="flex min-w-0 flex-wrap items-center justify-between gap-3 border-b border-base-300 py-3 last:border-0"
      >
        <div class="min-w-0 space-y-1">
          <strong>{{ proof.versionLabel }} · {{ proof.status }}</strong
          ><span>{{ proof.createdAt ? formatDateTime(proof.createdAt) : '' }}</span
          ><small v-if="proof.approverNote" class="block text-xs leading-5 text-base-content/60">{{
            proof.approverNote
          }}</small>
        </div>
      </div>
    </section>
  </div>
</template>
