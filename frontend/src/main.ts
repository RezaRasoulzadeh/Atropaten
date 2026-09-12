import { createApp } from 'vue'
import App from './app/AppShell.vue'
import './styles/fonts.css'
import './style.css'
import { installFormEnter } from './ui/enterSubmit'

createApp(App).mount('#app')
const removeFormEnter = installFormEnter(document)
if (import.meta.hot) import.meta.hot.dispose(removeFormEnter)
