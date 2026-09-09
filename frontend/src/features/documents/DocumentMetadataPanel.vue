<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import AppInput from '../../components/ui/AppInput.vue';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { ClipboardCheck, Eye, FilePlus2, Pencil, Plus, Trash2, X } from 'lucide-vue-next';
import StatusBadge from '../../components/ui/StatusBadge.vue';
import { metadataApi, type AttachmentRecord, type ProofRecord } from '../../api/metadata';
import { formatDateTime } from '../../utils/date';
import SelectField from '../../components/ui/SelectField.vue';
const props = withDefaults(
  defineProps<{ ownerType: 'order'; ownerId: string; protectedContext?: boolean }>(),
  { protectedContext: false },
);
const emit = defineEmits<{ notify: [message: string] }>();
const attachments = ref<AttachmentRecord[]>([]);
const attachmentThumbnails = ref<Record<string, string>>({});
const proofs = ref<ProofRecord[]>([]);
const loading = ref(true);
const fileInput = ref<HTMLInputElement | null>(null);
const selectedFiles = ref<File[]>([]);
const selectedPreviews = ref<{ file: File; url: string }[]>([]);
const dragging = ref(false);
const category = ref('artwork');
const notes = ref('');
const proofStatus = ref('Draft');
const proofVersion = ref('v1');
const proofNote = ref('');
const selectedAttachment = ref('');
const proofDialogOpen = ref(false);
const editingProofId = ref<string | null>(null);
const preview = ref<{ id: string; fileName: string; mimeType: string; url: string } | null>(null);
const previewLoading = ref(false);
const savingFile = ref(false);
async function load() {
  loading.value = true;
  try {
    const [nextAttachments, nextProofs] = await Promise.all([
      metadataApi.attachments(props.ownerType, props.ownerId),
      metadataApi.proofs(props.ownerType, props.ownerId),
    ]);
    attachments.value = nextAttachments;
    proofs.value = nextProofs;
    const thumbnails: Record<string, string> = {};
    await Promise.all(
      nextAttachments
        .filter((file) => file.mimeType.startsWith('image/'))
        .map(async (file) => {
          try {
            const content = await metadataApi.readAttachment(file.id);
            thumbnails[file.id] = `data:${content.mimeType};base64,${content.contentBase64}`;
          } catch {
            // Keep the file card usable when a legacy image path is unavailable.
          }
        }),
    );
    attachmentThumbnails.value = thumbnails;
  } catch (e) {
    emit('notify', String(e));
  } finally {
    loading.value = false;
  }
}
function chooseFiles(files: FileList | File[] | undefined) {
  if (!files) return;
  for (const file of Array.from(files)) {
    if (selectedFiles.value.some((current) => current.name === file.name && current.size === file.size && current.lastModified === file.lastModified)) continue;
    selectedFiles.value.push(file);
    if (file.type.startsWith('image/')) {
      selectedPreviews.value.push({ file, url: URL.createObjectURL(file) });
    }
  }
}
function onFileChange(event: Event) {
  const input = event.target as HTMLInputElement;
  chooseFiles(input.files ?? undefined);
  input.value = '';
}
function onDrop(event: DragEvent) {
  dragging.value = false;
  chooseFiles(event.dataTransfer?.files ?? undefined);
}
function removeSelectedFile(file: File) {
  selectedFiles.value = selectedFiles.value.filter((current) => current !== file);
  const preview = selectedPreviews.value.find((value) => value.file === file);
  if (preview) URL.revokeObjectURL(preview.url);
  selectedPreviews.value = selectedPreviews.value.filter((value) => value.file !== file);
}
function clearSelectedFiles() {
  for (const preview of selectedPreviews.value) URL.revokeObjectURL(preview.url);
  selectedFiles.value = [];
  selectedPreviews.value = [];
  if (fileInput.value) fileInput.value.value = '';
}
function formatFileSize(size: number) {
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  return `${(size / (1024 * 1024 * 1024)).toFixed(1)} GB`;
}
async function fileAsBase64(file: File) {
  const bytes = new Uint8Array(await file.arrayBuffer());
  let binary = '';
  const chunkSize = 0x8000;
  for (let offset = 0; offset < bytes.length; offset += chunkSize) {
    binary += String.fromCharCode(...bytes.subarray(offset, offset + chunkSize));
  }
  return btoa(binary);
}
async function addFile() {
return runAction(async () => {
  if (!selectedFiles.value.length) {
    emit('notify', 'Choose at least one file first');
    return;
  }
  try {
    for (const file of selectedFiles.value) {
      await metadataApi.importAttachment({
        ownerType: props.ownerType,
        ownerId: props.ownerId,
        fileName: file.name,
        mimeType: file.type || 'application/octet-stream',
        contentBase64: await fileAsBase64(file),
        category: category.value,
        notes: notes.value,
      });
    }
    const storedCount = selectedFiles.value.length;
    clearSelectedFiles();
    notes.value = '';
    await load();
    emit('notify', `${storedCount} file${storedCount === 1 ? '' : 's'} stored in attachments`);
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
async function previewFile(file: AttachmentRecord) {
  previewLoading.value = true;
  try {
    const cachedUrl = attachmentThumbnails.value[file.id];
    if (cachedUrl) {
      preview.value = { id: file.id, fileName: file.fileName, mimeType: file.mimeType, url: cachedUrl };
      return;
    }
    const content = await metadataApi.readAttachment(file.id);
    preview.value = {
      id: file.id,
      fileName: content.fileName,
      mimeType: content.mimeType,
      url: `data:${content.mimeType};base64,${content.contentBase64}`,
    };
  } catch (e) {
    reportError(e);
  } finally {
    previewLoading.value = false;
  }
}
async function saveAttachment() {
  if (!preview.value) return;
  savingFile.value = true;
  try {
    const saved = await metadataApi.saveAttachment(preview.value.id);
    if (saved) emit('notify', 'Attachment saved successfully');
  } catch (e) {
    reportError(e);
  } finally {
    savingFile.value = false;
  }
}
function closePreview() {
  preview.value = null;
}
async function saveProof() {
return runAction(async () => {
  try {
    const editing = Boolean(editingProofId.value);
    if (editingProofId.value) {
      await metadataApi.transitionProof(
        props.ownerType,
        props.ownerId,
        editingProofId.value,
        proofStatus.value,
        proofNote.value,
      );
    } else {
      await metadataApi.createProof({
        ownerType: props.ownerType,
        ownerId: props.ownerId,
        attachmentId: selectedAttachment.value,
        status: proofStatus.value,
        versionLabel: proofVersion.value,
        approverNote: proofNote.value,
        internalNote: '',
      });
    }
    proofNote.value = '';
    await load();
    closeProofDialog();
    emit('notify', editing ? 'Proof status updated' : 'Proof version recorded');
  } catch (e) {
reportError(e);
  }

});
}
function proofsForAttachment(attachmentId: string) {
  return proofs.value.filter((proof) => proof.attachmentId === attachmentId);
}
function openProofDialog(attachmentId = '', proof?: ProofRecord) {
  editingProofId.value = proof?.id ?? null;
  selectedAttachment.value = attachmentId || proof?.attachmentId || '';
  proofVersion.value = proof?.versionLabel || 'v1';
  proofStatus.value = proof?.status || 'Draft';
  proofNote.value = proof?.approverNote || '';
  proofDialogOpen.value = true;
}
function closeProofDialog() {
  proofDialogOpen.value = false;
  editingProofId.value = null;
}
onMounted(load);
onBeforeUnmount(clearSelectedFiles);
</script>
<template>
  <div class="min-w-0 space-y-4">
    <section class="min-w-0 overflow-hidden rounded-box border border-base-300 bg-base-100">
      <header class="flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-base-300 p-4">
        <div class="flex min-w-0 items-start gap-3">
          <div class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary">
            <FilePlus2 :size="18" aria-hidden="true" />
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-base font-semibold">Files & proofs</h2>
              <span class="badge badge-sm">{{ attachments.length }} files</span>
              <span class="badge badge-ghost badge-sm">{{ proofs.length }} proofs</span>
            </div>
            <p class="mt-1 text-xs text-base-content/60">
              Keep artwork and proof versions connected in one review-ready workspace.
            </p>
          </div>
        </div>
        <div class="flex flex-wrap items-center justify-end gap-2">
          <span class="text-xs text-base-content/50">Original bytes preserved</span>
          <button class="btn btn-primary btn-sm gap-2" type="button" @click="openProofDialog()" :disabled="busy">
            <ClipboardCheck :size="14" aria-hidden="true" />
            Record proof
          </button>
        </div>
      </header>

      <div class="grid min-w-0 gap-4 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(15rem,0.65fr)]">
        <div class="min-w-0 space-y-3">
          <div>
            <h3 class="text-sm font-semibold">Add attachment</h3>
            <p class="mt-1 text-xs text-base-content/60">Drop a file here or browse your computer. It will be copied into Atropaten’s managed attachments folder.</p>
          </div>
          <div class="grid min-w-0 gap-3 sm:grid-cols-2">
            <div
              class="group relative flex min-h-44 cursor-pointer flex-col items-center justify-center gap-2 rounded-box border border-dashed p-4 text-center transition-colors sm:col-span-2"
              :class="dragging ? 'border-primary bg-primary/10' : 'border-base-300 bg-base-200/25 hover:border-primary/60 hover:bg-primary/5'"
              role="button"
              tabindex="0"
              @click="fileInput?.click()"
              @keydown.enter="fileInput?.click()"
              @dragover.prevent="dragging = true"
              @dragleave.prevent="dragging = false"
              @drop.prevent="onDrop"
            >
              <input
                ref="fileInput"
                class="sr-only"
                type="file"
                multiple
                @click.stop
                @change="onFileChange"
              />
              <div class="grid size-10 place-items-center rounded-full bg-primary/10 text-primary transition-colors group-hover:bg-primary/15">
                <FilePlus2 :size="20" aria-hidden="true" />
              </div>
              <template v-if="selectedFiles.length">
                <strong class="text-sm">{{ selectedFiles.length }} file{{ selectedFiles.length === 1 ? '' : 's' }} selected</strong>
                <span class="text-xs text-base-content/55">Ready to store in the order attachments folder</span>
                <div class="grid w-full min-w-0 gap-2 text-start sm:grid-cols-2">
                  <div
                    v-for="file in selectedFiles"
                    :key="`${file.name}-${file.lastModified}`"
                    class="flex min-w-0 items-center gap-2 rounded-box border border-base-300 bg-base-100/80 p-2"
                  >
                    <img
                      v-if="selectedPreviews.find((preview) => preview.file === file)"
                      :src="selectedPreviews.find((preview) => preview.file === file)?.url"
                      :alt="file.name"
                      class="size-10 shrink-0 rounded object-cover"
                    />
                    <div v-else class="grid size-10 shrink-0 place-items-center rounded bg-base-200 text-primary">
                      <FilePlus2 :size="17" aria-hidden="true" />
                    </div>
                    <div class="min-w-0 flex-1">
                      <strong class="block truncate text-xs">{{ file.name }}</strong>
                      <span class="block text-[11px] text-base-content/55">{{ formatFileSize(file.size) }}</span>
                    </div>
                    <button
                      class="btn btn-outline btn-error btn-square btn-xs shrink-0"
                      type="button"
                      :aria-label="`Remove ${file.name}`"
                      title="Remove from selection"
                      @click.stop="removeSelectedFile(file)"
                    >
                      ×
                    </button>
                  </div>
                </div>
                <span class="text-xs text-primary">Drop more files or use Browse files to add to this selection</span>
              </template>
              <template v-else>
                <strong class="text-sm">Drop your file here</strong>
                <span class="text-xs text-base-content/55">or click Browse file to choose it</span>
              </template>
              <button class="btn btn-outline btn-sm mt-1" type="button" @click.stop="fileInput?.click()">
                {{ selectedFiles.length ? 'Browse more files' : 'Browse files' }}
              </button>
            </div>
            <SelectField
              v-model="category"
              label="Category"
              :options="[
                { label: 'Artwork', value: 'artwork' },
                { label: 'Proof', value: 'proof' },
                { label: 'Reference', value: 'reference' },
                { label: 'Other', value: 'other' },
              ]"
            />
            <FormField class="sm:col-span-2" label="Storage location"><AppInput
              class="input w-full min-w-0"
              :model-value="selectedFiles.length ? 'Atropaten managed attachments folder' : 'Choose files to store them here'"
              readonly
            /></FormField>
            <FormField class="sm:col-span-2" label="Notes"><AppInput
              class="input w-full min-w-0"
              v-model="notes"
              placeholder="Optional context for the team"
            /></FormField>
          </div>
          <button class="btn btn-primary btn-sm w-fit gap-2" type="button" @click="addFile" :disabled="busy || !selectedFiles.length">
            <Plus :size="14" aria-hidden="true" />
            {{ busy ? 'Storing…' : 'Store attachment' }}
          </button>
        </div>

        <div class="min-w-0 rounded-box border border-dashed border-base-300 bg-base-200/30 p-4 text-xs text-base-content/60">
          <strong class="block text-sm text-base-content/80">Managed file storage</strong>
          <p class="mt-2 leading-5">The original file is stored without resizing or compression, so it can be used later at the same quality.</p>
          <p class="mt-2 leading-5">The app keeps the file name, type, size, and checksum with this order for reliable backup and restore.</p>
        </div>
      </div>

      <div class="border-t border-base-300 p-4">
        <LoadingState v-if="loading" label="Loading attachments…" />
        <EmptyState
          v-else-if="!attachments.length && !proofs.length"
          compact
          title="No attachment references"
          description="Files and proofs linked to this record will appear here."
        >
          <template #icon><FilePlus2 :size="21" aria-hidden="true" /></template>
        </EmptyState>
        <div v-else class="grid min-w-0 gap-2">
          <div
            v-for="file in attachments"
            :key="file.id"
            class="flex min-w-0 flex-wrap items-start justify-between gap-3 rounded-box border border-base-300 bg-base-200/35 p-3"
          >
            <button
              class="group grid size-14 shrink-0 place-items-center overflow-hidden rounded-box border border-base-300 bg-base-200 text-primary"
              type="button"
              :aria-label="`Preview ${file.fileName}`"
              title="Preview file"
              @click="previewFile(file)"
            >
              <img
                v-if="attachmentThumbnails[file.id]"
                class="size-full object-cover transition-transform group-hover:scale-105"
                :src="attachmentThumbnails[file.id]"
                :alt="file.fileName"
              />
              <FilePlus2 v-else :size="21" aria-hidden="true" />
            </button>
            <div class="min-w-0">
              <div class="flex min-w-0 flex-wrap items-center gap-2">
                <strong class="truncate text-sm">{{ file.fileName }}</strong>
                <span class="badge badge-ghost badge-sm capitalize">{{ file.category }}</span>
              </div>
              <p class="mt-1 break-all text-xs text-base-content/60">{{ file.path }}</p>
              <span class="mt-1 block text-[11px] text-base-content/45">
                {{ file.mimeType || 'File' }}<template v-if="file.sizeBytes"> · {{ formatFileSize(file.sizeBytes) }}</template>
              </span>
              <small v-if="file.notes" class="mt-1 block text-xs leading-5 text-base-content/55">{{ file.notes }}</small>
            </div>
            <div class="flex shrink-0 items-center gap-1">
              <button
                class="btn btn-ghost btn-sm gap-1"
                type="button"
                @click="openProofDialog(file.id)"
                :disabled="busy"
              >
                <ClipboardCheck :size="14" aria-hidden="true" />
                Proof
              </button>
              <button
                class="btn btn-ghost btn-sm gap-1"
                type="button"
                @click="previewFile(file)"
                :disabled="previewLoading"
              >
                <Eye :size="14" aria-hidden="true" />
                {{ previewLoading ? 'Opening…' : 'Preview' }}
              </button>
              <button
                class="btn btn-outline btn-error btn-square btn-sm"
                aria-label="Remove attachment metadata"
                title="Remove attachment"
                type="button"
                @click="removeFile(file.id)"
                :disabled="busy || props.protectedContext"
              >
                <Trash2 :size="14" aria-hidden="true" />
              </button>
            </div>
            <div
              v-if="proofsForAttachment(file.id).length"
              class="mt-3 basis-full rounded-box border border-base-300/80 bg-base-100/60 p-3"
            >
              <div class="mb-2 flex items-center gap-2 text-xs text-base-content/60">
                <ClipboardCheck :size="14" class="text-primary" aria-hidden="true" />
                <span class="font-medium text-base-content/75">Proof history for this file</span>
              </div>
              <div class="grid gap-2">
                <div
                  v-for="proof in proofsForAttachment(file.id)"
                  :key="proof.id"
                  class="flex min-w-0 flex-wrap items-center justify-between gap-2 rounded-box bg-base-200/45 px-3 py-2"
                >
                  <div class="min-w-0">
                    <div class="flex flex-wrap items-center gap-2">
                      <strong class="text-xs">{{ proof.versionLabel }}</strong>
                      <StatusBadge :label="proof.status" :tone="proof.status === 'Approved' ? 'green' : proof.status === 'Rejected' ? 'red' : 'slate'" />
                    </div>
                    <span class="mt-1 block text-[11px] text-base-content/50">{{ proof.createdAt ? formatDateTime(proof.createdAt) : 'No date recorded' }}</span>
                  </div>
                  <small v-if="proof.approverNote" class="max-w-md text-xs text-base-content/60">{{ proof.approverNote }}</small>
                  <button
                    class="btn btn-ghost btn-square btn-xs shrink-0"
                    type="button"
                    aria-label="Update proof status"
                    title="Update proof status"
                    @click="openProofDialog(file.id, proof)"
                    :disabled="busy"
                  >
                    <Pencil :size="13" aria-hidden="true" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div
          v-if="proofs.some((proof) => !proof.attachmentId)"
          class="rounded-box border border-dashed border-base-300 bg-base-200/25 p-3"
        >
          <div class="mb-2 flex items-center justify-between gap-2">
            <div class="flex items-center gap-2 text-xs text-base-content/60">
              <ClipboardCheck :size="14" class="text-primary" aria-hidden="true" />
              <span class="font-medium text-base-content/75">Unlinked proof history</span>
            </div>
            <button class="btn btn-ghost btn-xs" type="button" @click="openProofDialog()">Add proof</button>
          </div>
          <div class="grid gap-2">
            <div
              v-for="proof in proofs.filter((item) => !item.attachmentId)"
              :key="proof.id"
              class="flex min-w-0 flex-wrap items-center justify-between gap-2 rounded-box bg-base-100/70 px-3 py-2"
            >
              <div class="min-w-0">
                <div class="flex flex-wrap items-center gap-2">
                  <strong class="text-xs">{{ proof.versionLabel }}</strong>
                  <StatusBadge :label="proof.status" :tone="proof.status === 'Approved' ? 'green' : proof.status === 'Rejected' ? 'red' : 'slate'" />
                </div>
                <span class="mt-1 block text-[11px] text-base-content/50">{{ proof.createdAt ? formatDateTime(proof.createdAt) : 'No date recorded' }}</span>
              </div>
              <small v-if="proof.approverNote" class="max-w-md text-xs text-base-content/60">{{ proof.approverNote }}</small>
              <button
                class="btn btn-ghost btn-square btn-xs shrink-0"
                type="button"
                aria-label="Update proof status"
                title="Update proof status"
                @click="openProofDialog('', proof)"
                :disabled="busy"
              >
                <Pencil :size="13" aria-hidden="true" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <div
      v-if="proofDialogOpen"
      class="fixed inset-0 z-40 flex items-center justify-center bg-black/60 p-4"
      role="dialog"
      aria-modal="true"
      :aria-label="editingProofId ? 'Update proof status' : 'Record proof version'"
      @click.self="closeProofDialog"
      @keydown.esc="closeProofDialog"
      tabindex="-1"
    >
      <div class="w-full max-w-xl overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-2xl">
        <header class="flex items-start justify-between gap-3 border-b border-base-300 p-4">
          <div class="flex min-w-0 items-start gap-3">
            <div class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary">
              <ClipboardCheck :size="18" aria-hidden="true" />
            </div>
            <div class="min-w-0">
              <h2 class="text-base font-semibold">{{ editingProofId ? 'Update proof status' : 'Record proof version' }}</h2>
              <p class="mt-1 text-xs text-base-content/60">
                {{ editingProofId ? 'Update the review state or note for this proof.' : selectedAttachment ? 'This proof will appear under the selected file.' : 'Link this proof to a file or keep it unlinked for now.' }}
              </p>
            </div>
          </div>
          <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Close proof dialog" @click="closeProofDialog">
            <X :size="17" aria-hidden="true" />
          </button>
        </header>
        <div class="grid gap-4 p-4">
          <div class="grid gap-3 sm:grid-cols-2">
            <FormField label="Version / label"><AppInput
              class="input w-full min-w-0"
              v-model="proofVersion"
              placeholder="e.g. v2"
              :readonly="!!editingProofId"
            /></FormField>
            <SelectField
              v-model="proofStatus"
              label="Status"
              :options="
                ['Draft', 'Ready', 'Waiting Customer Approval', 'Approved', 'Rejected'].map(
                  (value) => ({ label: value, value }),
                )
              "
            />
            <SelectField
              class="sm:col-span-2"
              v-model="selectedAttachment"
              label="Related attachment"
              :options="[
                { label: 'No related attachment', value: '' },
                ...attachments.map((file) => ({ label: file.fileName, value: file.id })),
              ]"
              :disabled="!!editingProofId"
            />
            <FormField class="sm:col-span-2" label="Approval / rejection note"><AppInput
              class="input w-full min-w-0"
              v-model="proofNote"
              placeholder="Optional note for this version"
            /></FormField>
          </div>
          <div class="flex flex-wrap justify-end gap-2 border-t border-base-300 pt-4">
            <button class="btn btn-outline btn-sm" type="button" @click="closeProofDialog">Cancel</button>
            <button class="btn btn-primary btn-sm gap-2" type="button" @click="saveProof" :disabled="busy">
              <Plus :size="14" aria-hidden="true" />
              {{ busy ? 'Saving…' : editingProofId ? 'Save changes' : 'Record proof' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="preview"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
      role="dialog"
      aria-modal="true"
      :aria-label="`Preview ${preview.fileName}`"
      @click.self="closePreview"
      @keydown.esc="closePreview"
      tabindex="-1"
    >
      <div class="flex max-h-full w-full max-w-6xl min-w-0 flex-col overflow-hidden rounded-box border border-base-300 bg-base-100 shadow-2xl">
        <header class="flex min-w-0 items-center justify-between gap-3 border-b border-base-300 px-4 py-3">
          <div class="min-w-0">
            <strong class="block truncate text-sm">{{ preview.fileName }}</strong>
            <span class="block text-xs text-base-content/55">{{ preview.mimeType }}</span>
          </div>
          <div class="flex shrink-0 items-center gap-2">
            <button class="btn btn-outline btn-sm" type="button" :disabled="savingFile" @click="saveAttachment">
              {{ savingFile ? 'Saving…' : 'Save as…' }}
            </button>
            <button class="btn btn-ghost btn-square btn-sm" type="button" aria-label="Close preview" @click="closePreview">
              <X :size="17" aria-hidden="true" />
            </button>
          </div>
        </header>
        <div class="flex min-h-0 flex-1 items-center justify-center overflow-auto bg-base-200 p-4">
          <img
            v-if="preview.mimeType.startsWith('image/')"
            class="max-h-[calc(100vh-10rem)] max-w-full object-contain"
            :src="preview.url"
            :alt="preview.fileName"
          />
          <video
            v-else-if="preview.mimeType.startsWith('video/')"
            class="max-h-[calc(100vh-10rem)] max-w-full"
            :src="preview.url"
            controls
          ></video>
          <audio v-else-if="preview.mimeType.startsWith('audio/')" class="w-full max-w-xl" :src="preview.url" controls></audio>
          <iframe
            v-else-if="preview.mimeType === 'application/pdf' || preview.mimeType.startsWith('text/')"
            class="h-[calc(100vh-10rem)] w-full rounded-box bg-base-100"
            :src="preview.url"
            :title="`Preview ${preview.fileName}`"
          ></iframe>
          <div v-else class="max-w-md text-center">
            <FilePlus2 :size="34" class="mx-auto text-base-content/40" aria-hidden="true" />
            <p class="mt-3 text-sm text-base-content/70">This file type cannot be previewed here.</p>
            <button class="btn btn-primary btn-sm mt-4" type="button" :disabled="savingFile" @click="saveAttachment">
              {{ savingFile ? 'Saving…' : 'Save original file' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
