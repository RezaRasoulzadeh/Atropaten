import { createApp } from 'vue'
import App from './app/AppShell.vue'
import './styles/fonts.css'
import './style.css'
import { installFormEnter } from './ui/enterSubmit'
import { i18n, translateUi } from './i18n'

const app = createApp(App)
app.config.globalProperties.$ui = translateUi
app.use(i18n).mount('#app')
const removeFormEnter = installFormEnter(document)
if (import.meta.hot) import.meta.hot.dispose(removeFormEnter)
