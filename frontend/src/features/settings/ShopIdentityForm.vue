<script setup lang="ts">
import type { ShopSettingsRecord } from '../../api/reports'
import FormField from '../../components/ui/FormField.vue'
import FormSection from '../../components/ui/FormSection.vue'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
const form = defineModel<ShopSettingsRecord>({ required: true })
</script>
<template>
<div class="space-y-4">
<FormSection title="Shop identity">
<FormField label="Shop name" required><AppInput v-model="form.shopName" required /></FormField>
<FormField label="Subtitle"><AppInput v-model="form.shopSubtitle" /></FormField>
<FormField label="Registration ID"><AppInput v-model="form.registrationId" /></FormField>
<FormField label="Tax ID"><AppInput v-model="form.taxId" /></FormField>
</FormSection>
<FormSection title="Contact">
<FormField label="Phone"><AppInput v-model="form.phone" type="tel" /></FormField>
<FormField label="Email"><AppInput v-model="form.email" type="email" /></FormField>
<FormField label="Website"><AppInput v-model="form.website" /></FormField>
<FormField label="Logo path"><AppInput v-model="form.logoPath" /></FormField>
<FormField label="Address" class="sm:col-span-2"><AppTextarea v-model="form.address" rows="2" /></FormField>
</FormSection>
<FormSection title="Printed documents">
<FormField label="Default footer"><AppTextarea v-model="form.documentFooter" rows="3" /></FormField>
<FormField label="Default notes"><AppTextarea v-model="form.documentNotes" rows="3" /></FormField>
</FormSection>
</div>
</template>
