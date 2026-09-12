<script setup lang="ts">
import LoadingState from '../../components/ui/LoadingState.vue'
import EmptyState from '../../components/ui/EmptyState.vue'
import FormField from '../../components/ui/FormField.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import AppInput from '../../components/ui/AppInput.vue';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import { Eye, FilePlus2, Plus, Trash2, X } from 'lucide-vue-next';
import { metadataApi, type AttachmentRecord } from '../../api/metadata';
import SelectField from '../../components/ui/SelectField.vue';
const props = withDefaults(
  defineProps<{ ownerType: 'order'; ownerId: string; protectedContext?: boolean }>(),
  { protectedContext: false },
);
const emit = defineEmits<{ notify: [message: string] }>();
const attachments = ref<AttachmentRecord[]>([]);
const attachmentThumbnails = ref<Record<string, string>>({});
const loading = ref(true);
const fileInput = ref<HTMLInputElement | null>(null);
const selectedFiles = ref<File[]>([]);
const selectedPreviews = ref<{ file: File; url: string }[]>([]);
const dragging = ref(false);
const category = ref('artwork');
const notes = ref('');
const preview = ref<{ id: string; fileName: string; mimeType: string; url: string } | null>(null);
const previewLoading = ref(false);
const savingFile = ref(false);
async function load() {
  loading.value = true;
  try {
    const nextAttachments = await metadataApi.attachments(props.ownerType, props.ownerId);
    attachments.value = nextAttachments;
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
onMounted(load);
onBeforeUnmount(clearSelectedFiles);
</script>
<template>
  <div class="min-w-0">
    <section class="min-w-0 overflow-hidden">
      <header class="flex min-w-0 flex-wrap items-start justify-between gap-3 border-b border-base-300 pb-4">
        <div class="flex min-w-0 items-start gap-3">
          <div class="grid size-9 shrink-0 place-items-center rounded-box bg-primary/10 text-primary">
            <FilePlus2 :size="18" aria-hidden="true" />
          </div>
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="text-base font-semibold">Files</h2>
              <span class="badge badge-sm">{{ attachments.length }} files</span>
            </div>
            <p class="mt-1 text-xs text-base-content/60">
              Add artwork and reference files to this order.
            </p>
          </div>
        </div>
      </header>

      <div class="min-w-0 pt-4">
        <div data-enter-scope class="min-w-0 space-y-3">
          <div>
            <h3 class="text-sm font-semibold">Add attachment</h3>
            <p class="mt-1 text-xs text-base-content/60">Drop a file here or browse your computer. New files use the folder configured in Settings.</p>
          </div>
          <div class="grid min-w-0 gap-3 sm:grid-cols-2">
            <div
              class="group relative flex min-h-32 cursor-pointer flex-col items-center justify-center gap-2 rounded-box border border-dashed p-3 text-center transition-colors sm:col-span-2"
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
                { label: 'Reference', value: 'reference' },
                { label: 'Other', value: 'other' },
              ]"
            />
            <FormField label="Notes"><AppInput
              class="input w-full min-w-0"
              v-model="notes"
              placeholder="Optional context for the team"
            /></FormField>
          </div>
          <button class="btn btn-primary btn-sm w-fit gap-2" type="button" data-enter-submit @click="addFile" :disabled="busy || !selectedFiles.length">
            <Plus :size="14" aria-hidden="true" />
            {{ busy ? 'Storing…' : 'Store attachment' }}
          </button>
        </div>
      </div>

      <div class="mt-4 border-t border-base-300 pt-4">
        <LoadingState v-if="loading" label="Loading attachments…" />
        <EmptyState
          v-else-if="!attachments.length"
          compact
          title="No files yet"
          description="Files linked to this order will appear here."
        >
          <template #icon><FilePlus2 :size="21" aria-hidden="true" /></template>
        </EmptyState>
        <div v-else class="grid min-w-0 gap-2">
          <div
            v-for="file in attachments"
            :key="file.id"
            class="grid min-w-0 grid-cols-[2.5rem_minmax(0,1fr)] items-center gap-2 rounded-box border border-base-300 bg-base-200/35 p-2 sm:grid-cols-[2.5rem_minmax(0,1fr)_auto] sm:gap-3"
          >
            <button
              class="group grid size-10 shrink-0 place-items-center overflow-hidden rounded-box border border-base-300 bg-base-200 text-primary"
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
                <strong class="min-w-0 truncate text-sm" :title="file.fileName">{{ file.fileName }}</strong>
                <span class="badge badge-ghost badge-sm capitalize">{{ file.category }}</span>
              </div>
              <p class="mt-0.5 truncate text-[11px] text-base-content/55" :title="file.path">{{ file.path }}</p>
              <span class="mt-0.5 block text-[11px] text-base-content/45">
                {{ file.mimeType || 'File' }}<template v-if="file.sizeBytes"> · {{ formatFileSize(file.sizeBytes) }}</template>
              </span>
              <small v-if="file.notes" class="mt-0.5 block truncate text-[11px] text-base-content/55" :title="file.notes">{{ file.notes }}</small>
            </div>
            <div class="col-start-2 flex shrink-0 items-center gap-1 sm:col-start-auto">
              <button
                class="btn btn-ghost btn-xs gap-1 px-2"
                type="button"
                @click="previewFile(file)"
                :disabled="previewLoading"
              >
                <Eye :size="14" aria-hidden="true" />
                {{ previewLoading ? 'Opening…' : 'Preview' }}
              </button>
              <button
                class="btn btn-outline btn-error btn-square btn-xs"
                aria-label="Remove attachment metadata"
                title="Remove attachment"
                type="button"
                @click="removeFile(file.id)"
                :disabled="busy || props.protectedContext"
              >
                <Trash2 :size="14" aria-hidden="true" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </section>

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
