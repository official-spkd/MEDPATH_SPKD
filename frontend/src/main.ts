import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import './assets/base.css'

const app = createApp(App)
app.config.errorHandler = (err, _inst, info) => {
  // eslint-disable-next-line no-console
  console.error('[VUE ERR]', info, (err as Error)?.message, (err as Error)?.stack)
}
app.use(createPinia()).use(router).mount('#app')
