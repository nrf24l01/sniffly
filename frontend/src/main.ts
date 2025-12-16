import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { useThemeStore } from './stores/theme'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart } from 'echarts/charts'
import { TitleComponent, TooltipComponent, GridComponent, LegendComponent } from 'echarts/components'

import App from './App.vue'
import router from './router'

const app = createApp(App)

use([CanvasRenderer, BarChart, LineChart, TitleComponent, TooltipComponent, GridComponent, LegendComponent])

const pinia = createPinia()

app.use(pinia)
app.use(router)

app.component('VChart', VChart)

const themeStore = useThemeStore()
themeStore.initTheme()

app.mount('#app')
