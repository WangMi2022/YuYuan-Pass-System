import VCharts from 'vue-echarts'
import { use } from 'echarts/core'
import {
  AriaComponent,
  GraphicComponent,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

const commonComponents = [
  AriaComponent,
  GraphicComponent,
  GridComponent,
  LegendComponent,
  TitleComponent,
  TooltipComponent,
  CanvasRenderer
]

export const registerChart = (chart) => {
  use([...commonComponents, chart])
}

export { VCharts }
