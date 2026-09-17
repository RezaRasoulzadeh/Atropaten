import type { translateUi } from './index';

declare module 'vue' {
  interface ComponentCustomProperties {
    $ui: typeof translateUi;
  }
}

export {};
