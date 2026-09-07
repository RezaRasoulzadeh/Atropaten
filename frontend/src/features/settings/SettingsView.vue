<script setup lang="ts">
import ShopIdentityForm from './ShopIdentityForm.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import InlineAlert from '../../components/ui/InlineAlert.vue'
import MasterDetail from '../../components/layout/MasterDetail.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

import FormGrid from '../../components/ui/FormGrid.vue';
import WorkspaceHeader from '../../components/layout/WorkspaceHeader.vue';
import AppInput from '../../components/ui/AppInput.vue';
import FormField from '../../components/ui/FormField.vue';
import AppPanel from '../../components/layout/AppPanel.vue';
import { onMounted, ref } from 'vue';
import WorkspaceStickyStack from '../../components/layout/WorkspaceStickyStack.vue';
import {
  reportsApi,
  type BackupInfoRecord,
  type DataPathsRecord,
  type ShopSettingsRecord,
} from '../../api/reports';
import { confirmAction, normalizeError } from '../../ui/feedback';
const emit = defineEmits<{ notify: [message: string] }>();
const form = ref<ShopSettingsRecord>({
  shopName: '',
  shopSubtitle: '',
  phone: '',
  address: '',
  email: '',
  website: '',
  registrationId: '',
  taxId: '',
  logoPath: '',
  documentFooter: '',
  documentNotes: '',
});
const loading = ref(true);
const loadError = ref('');
const backupError = ref('');
const backupBusy = ref(false);
const backupPath = ref('');
const paths = ref<DataPathsRecord | null>(null);
const lastBackup = ref<BackupInfoRecord | null>(null);
onMounted(async () => {
  try {
    const [settings, dataPaths] = await Promise.all([
      reportsApi.settings(),
      reportsApi.dataPaths(),
    ]);
    form.value = settings;
    paths.value = dataPaths;
    try {
      lastBackup.value = await reportsApi.lastBackup();
    } catch (error) { backupError.value = normalizeError(error).message; }
  } catch (e) {
reportError(e);
    loadError.value = normalizeError(e).message;
  } finally {
    loading.value = false;
  }
});
async function save() {
return runAction(async () => {
  try {
    await reportsApi.saveSettings(form.value);
    emit('notify', 'Document identity saved.');
  } catch (e) {
reportError(e);
  }

});
}
async function createBackup() {
return runAction(async () => {
  backupBusy.value = true;
  try {
    lastBackup.value = await reportsApi.createBackup();
    backupPath.value = lastBackup.value.path;
    emit('notify', 'Backup created and verified.');
  } catch (e) {
reportError(e);
  } finally {
    backupBusy.value = false;
  }

});
}
async function chooseBackup() {
return runAction(async () => {
  try {
    const path = await reportsApi.selectBackupFile();
    if (path) backupPath.value = path;
  } catch (e) {
reportError(e);
  }

});
}
async function verifyBackup() {
return runAction(async () => {
  if (!backupPath.value) return;
  backupBusy.value = true;
  try {
    lastBackup.value = await reportsApi.verifyBackup(backupPath.value);
    emit('notify', 'Backup is valid.');
  } catch (e) {
reportError(e);
  } finally {
    backupBusy.value = false;
  }

});
}
async function restoreBackup() {
return runAction(async () => {
  if (
    !backupPath.value ||
    !(await confirmAction({
      title: 'Restore backup',
      message:
        'Restore this backup? The current application data will be replaced after validation.',
      confirmLabel: 'Restore backup',
      danger: true,
    }))
  )
    return;
  backupBusy.value = true;
  try {
    lastBackup.value = await reportsApi.restoreBackup(backupPath.value);
    emit('notify', 'Backup restored successfully.');
  } catch (e) {
reportError(e);
  } finally {
    backupBusy.value = false;
  }

});
}
</script>
<template><div class="space-y-4">
<WorkspaceStickyStack><WorkspaceHeader eyebrow="Insights & setup" title="Shop settings" description="Document identity, contact details and data safety."><button class="btn btn-primary" type="submit" form="shop-settings" :disabled="loading || busy || !!loadError">Save settings</button></WorkspaceHeader></WorkspaceStickyStack>
<LoadingState v-if="loading" label="Loading settings…" /><InlineAlert v-else-if="loadError" :message="loadError" />
<MasterDetail v-else wide>
<AppPanel title="Document identity"><form id="shop-settings" @submit.prevent="save"><ShopIdentityForm v-model="form" /></form></AppPanel>
<AppPanel title="Backup & restore" subtitle="Backups include the database and managed files.">
<dl v-if="paths" class="space-y-3 text-sm"><div v-for="item in [{label:'Data location',value:paths.root},{label:'Database',value:paths.database},{label:'Version / schema',value:paths.applicationVersion+' · v'+paths.schemaVersion},{label:'Backups folder',value:paths.backups}]" :key="item.label"><dt class="text-xs text-base-content/60">{{item.label}}</dt><dd class="mt-1 wrap-anywhere">{{item.value}}</dd></div></dl>
<InlineAlert v-if="backupError" tone="warning" :message="backupError" />
<div class="flex flex-wrap gap-2 border-t border-base-300 pt-3"><button class="btn btn-primary" :disabled="backupBusy || busy" @click="createBackup">Create backup</button><button class="btn btn-outline" :disabled="backupBusy || busy" @click="chooseBackup">Choose backup</button></div>
<FormField label="Selected backup"><AppInput v-model="backupPath" placeholder="Path to a .zip backup" /></FormField>
<div class="flex flex-wrap gap-2"><button class="btn btn-outline" :disabled="backupBusy || busy || !backupPath" @click="verifyBackup">Verify selected</button><button class="btn btn-ghost text-error" :disabled="backupBusy || busy || !backupPath" @click="restoreBackup">Restore selected</button></div>
<p v-if="lastBackup" class="text-xs leading-5 text-base-content/60 wrap-anywhere">{{lastBackup.path}} · schema v{{lastBackup.schemaVersion}} · {{lastBackup.managedFileCount}} managed files</p>
</AppPanel></MasterDetail></div></template>
