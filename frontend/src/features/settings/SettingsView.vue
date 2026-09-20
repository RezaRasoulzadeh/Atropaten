<script setup lang="ts">
import ShopIdentityForm from './ShopIdentityForm.vue'
import LoadingState from '../../components/ui/LoadingState.vue'
import {useWorkspaceActions, reportError} from '../../composables/useWorkspaceActions'
const {busy,runAction}=useWorkspaceActions()

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
import { confirmAction } from '../../ui/feedback';
const emit = defineEmits<{ notify: [message: string]; restored: []; 'settings-updated': [settings: ShopSettingsRecord] }>();
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
  backupDirectory: '',
  attachmentDirectory: '',
  monetaryRoundingStepRial: 1000,
});
const loading = ref(true);
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
    form.value = {...settings, backupDirectory: settings.backupDirectory ?? '', attachmentDirectory: settings.attachmentDirectory ?? ''};
    paths.value = dataPaths;
    try {
      lastBackup.value = await reportsApi.lastBackup();
    } catch (error) { reportError(error); }
  } catch (e) {
reportError(e);
  } finally {
    loading.value = false;
  }
});
async function save() {
return runAction(async () => {
  try {
    await reportsApi.saveSettings(form.value);
    paths.value = await reportsApi.dataPaths();
    emit('settings-updated', { ...form.value });
    emit('notify', 'Settings saved.');
  } catch (e) {
reportError(e);
  }

});
}
async function chooseBackupDirectory() {
return runAction(async () => {
  try {
    const path = await reportsApi.selectBackupDirectory();
    if (path) form.value.backupDirectory = path;
  } catch (e) {
reportError(e);
  }

});
}
async function chooseAttachmentDirectory() {
return runAction(async () => {
  try {
    const path = await reportsApi.selectAttachmentDirectory();
    if (path) {
      const persistedSettings = await reportsApi.settings();
      await reportsApi.saveSettings({...persistedSettings, attachmentDirectory: path});
      form.value.attachmentDirectory = path;
      paths.value = await reportsApi.dataPaths();
    }
  } catch (e) {
reportError(e);
  }

});
}
async function resetAttachmentDirectory() {
return runAction(async () => {
  try {
    const persistedSettings = await reportsApi.settings();
    await reportsApi.saveSettings({...persistedSettings, attachmentDirectory: ''});
    form.value.attachmentDirectory = '';
    paths.value = await reportsApi.dataPaths();
  } catch (e) {
reportError(e);
  }

});
}
async function createBackup() {
return runAction(async () => {
  backupBusy.value = true;
  try {
    // The directory picker updates the form immediately, but the Go backup
    // service reads the persisted setting. Save only this setting first so a
    // newly selected folder is the folder that actually receives the archive.
    const persistedSettings = await reportsApi.settings();
    if ((persistedSettings.backupDirectory ?? '').trim() !== form.value.backupDirectory.trim()) {
      await reportsApi.saveSettings({
        ...persistedSettings,
        backupDirectory: form.value.backupDirectory,
      });
      paths.value = await reportsApi.dataPaths();
    }
    lastBackup.value = await reportsApi.createBackup(form.value.backupDirectory);
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
    const [settings, dataPaths] = await Promise.all([
      reportsApi.settings(),
      reportsApi.dataPaths(),
    ]);
    form.value = {
      ...settings,
      backupDirectory: settings.backupDirectory ?? '',
      attachmentDirectory: settings.attachmentDirectory ?? '',
    };
    paths.value = dataPaths;
    emit('restored');
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
<WorkspaceStickyStack><WorkspaceHeader :show-breadcrumb="true" :eyebrow='$t("Insights & setup")' :title='$t("Shop settings")' :description='$t("Document identity, contact details and data safety.")' /></WorkspaceStickyStack>
<LoadingState v-if="loading" :label='$t("Loading settings…")' />
<form v-else id="shop-settings" class="grid min-w-0 items-start gap-4 xl:grid-cols-[minmax(0,1fr)_30rem]" @submit.prevent="save">
<AppPanel :title='$t("Document identity")'>
<ShopIdentityForm v-model="form" />
<div class="mt-4 flex justify-end border-t border-base-300 pt-4"><button class="btn btn-primary" type="submit" :disabled="loading || busy">{{ $t("Save settings") }}</button></div>
</AppPanel>
<AppPanel :title='$t("Backup & restore")' :subtitle='$t("Backups include the database and managed files.")'>
<dl v-if="paths" class="space-y-3 text-sm"><div v-for="item in [{label:'Data location',value:paths.root},{label:'Database',value:paths.database},{label:'Version / schema',value:paths.applicationVersion+' · v'+paths.schemaVersion},{label:'Backups folder',value:paths.backups}]" :key="item.label"><dt class="text-xs text-base-content/60">{{$ui(item.label)}}</dt><dd class="mt-1 wrap-anywhere">{{item.value}}</dd></div></dl>
<div class="space-y-3 border-t border-base-300 pt-3">
<div class="space-y-2"><div><h3 class="text-sm font-semibold">{{ $t("File storage") }}</h3><p class="mt-1 text-xs leading-5 text-base-content/60">{{ $t("Choose the global folder where new order attachments are stored.") }}</p></div><FormField :label='$t("Attachments directory")' :help='$t("Leave empty to use Atropaten’s application-managed folder.")'><div class="flex min-w-0 gap-2"><AppInput :model-value="form.attachmentDirectory || paths?.attachments || 'Application-managed attachments folder'" readonly /><button class="btn btn-outline shrink-0" type="button" :disabled="busy || backupBusy" @click="chooseAttachmentDirectory">{{ $t("Browse") }}</button><button v-if="form.attachmentDirectory" class="btn btn-ghost shrink-0" type="button" :disabled="busy || backupBusy" @click="resetAttachmentDirectory">{{ $t("Reset") }}</button></div></FormField></div>
<FormField :label='$t("Backup directory")' :help='$t("Leave empty to use the application-managed backups folder.")'><div class="flex min-w-0 gap-2"><AppInput v-model="form.backupDirectory" :placeholder='$t("Application-managed backup folder")' /><button class="btn btn-outline shrink-0" type="button" :disabled="busy || backupBusy" @click="chooseBackupDirectory">{{ $t("Browse") }}</button></div></FormField>
<div class="flex flex-wrap gap-2"><button class="btn btn-primary" type="button" :disabled="backupBusy || busy" @click="createBackup">{{ $t("Create backup") }}</button><button class="btn btn-outline" type="button" :disabled="backupBusy || busy" @click="chooseBackup">{{ $t("Choose backup") }}</button></div>
</div>
<div data-enter-scope class="space-y-3">
<FormField :label='$t("Selected backup")'><AppInput v-model="backupPath" :placeholder='$t("Path to a .zip backup")' /></FormField>
<div class="flex flex-wrap gap-2"><button data-enter-submit class="btn btn-outline" type="button" :disabled="backupBusy || busy || !backupPath" @click="verifyBackup">{{ $t("Verify selected") }}</button><button class="btn btn-ghost text-error" type="button" :disabled="backupBusy || busy || !backupPath" @click="restoreBackup">{{ $t("Restore selected") }}</button></div>
</div>
<p v-if="lastBackup" class="text-xs leading-5 text-base-content/60 wrap-anywhere">{{lastBackup.path}} {{ $t("· schema v") }}{{lastBackup.schemaVersion}} · {{lastBackup.managedFileCount}} {{ $t("managed files") }}</p>
</AppPanel>
</form></div></template>
