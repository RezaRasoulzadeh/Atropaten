<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';
import { confirmState, resolveConfirm } from '../../ui/feedback';
import { translateUi } from '../../i18n';
const dialog = ref<HTMLDialogElement>();
let previousFocus: HTMLElement | null = null;
watch(
  () => confirmState.open,
  async (open) => {
    await nextTick();
    if (open) {
      previousFocus = document.activeElement as HTMLElement;
      dialog.value?.showModal();
    } else {
      dialog.value?.close();
      if (previousFocus?.isConnected) previousFocus.focus();
    }
  },
);
</script>
<template>
  <dialog
    ref="dialog"
    class="modal"
    role="alertdialog"
    aria-modal="true"
    aria-labelledby="confirm-title"
    aria-describedby="confirm-message"
    @cancel.prevent="resolveConfirm(false)"
    @close="confirmState.open && resolveConfirm(false)"
  >
    <div class="modal-box border border-base-300 bg-base-100 shadow-none">
      <h2 id="confirm-title" class="text-lg font-semibold leading-7">{{ translateUi(confirmState.title) }}</h2>
      <p id="confirm-message" class="mt-3 text-sm leading-5 text-base-content/70">
        {{ translateUi(confirmState.message) }}
      </p>
      <div class="modal-action">
        <button autofocus class="btn btn-ghost" type="button" @click="resolveConfirm(false)">
          {{ translateUi(confirmState.cancelLabel) }}</button
        ><button
          class="btn"
          :class="confirmState.danger ? 'btn-error' : 'btn-primary'"
          type="button"
          @click="resolveConfirm(true)"
        >
          {{ translateUi(confirmState.confirmLabel) }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button :aria-label='$t("Cancel confirmation")' @click="resolveConfirm(false)">{{ $t("Close") }}</button>
    </form>
  </dialog>
</template>
