<script setup lang="ts">
import type { ShopSettingsRecord } from '../../api/reports'
import FormField from '../../components/ui/FormField.vue'
import FormSection from '../../components/ui/FormSection.vue'
import AppInput from '../../components/ui/AppInput.vue'
import AppTextarea from '../../components/ui/AppTextarea.vue'
import { formatLocalizedNumber } from '../../utils/number'
const form = defineModel<ShopSettingsRecord>({ required: true })
</script>
<template>
<div class="space-y-4">
<FormSection :title='$t("Shop identity")'>
<FormField :label='$t("Shop name")' required><AppInput v-model="form.shopName" required /></FormField>
<FormField :label='$t("Subtitle")'><AppInput v-model="form.shopSubtitle" /></FormField>
<FormField :label='$t("Registration ID")'><AppInput v-model="form.registrationId" /></FormField>
<FormField :label='$t("Tax ID")'><AppInput v-model="form.taxId" /></FormField>
</FormSection>
<FormSection :title='$t("Contact")'>
<FormField :label='$t("Phone")'><AppInput v-model="form.phone" type="tel" /></FormField>
<FormField :label='$t("Email")'><AppInput v-model="form.email" type="email" /></FormField>
<FormField :label='$t("Website")'><AppInput v-model="form.website" /></FormField>
<FormField :label='$t("Logo path")'><AppInput v-model="form.logoPath" /></FormField>
<FormField :label='$t("Address")' class="sm:col-span-2"><AppTextarea v-model="form.address" rows="2" /></FormField>
</FormSection>
<FormSection :title='$t("Printed documents")'>
<FormField :label='$t("Default footer")'><AppTextarea v-model="form.documentFooter" rows="3" /></FormField>
<FormField :label='$t("Default notes")'><AppTextarea v-model="form.documentNotes" rows="3" /></FormField>
</FormSection>
<FormSection :title='$t("Money calculation")'>
<FormField :label='$t("Round calculated amounts up to")' :help='$t("Stored in Rial. For example, 100 Toman equals 1,000 Rial. Enter 1 for no effective rounding.")'><div class="grid gap-2 sm:grid-cols-2"><AppInput v-model.number="form.monetaryRoundingStepRial" type="number" min="1" step="1" /><div class="rounded border border-base-300 px-3 py-2 text-sm text-base-content/70">{{ $ui(form.monetaryRoundingStepRial ? `${formatLocalizedNumber(form.monetaryRoundingStepRial / 10)} Toman` : 'Enter a positive Rial step') }}</div></div></FormField>
</FormSection>
</div>
</template>
