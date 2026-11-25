/*
 * 应用核心插件入口
 */
import { register } from './global'

export default {
  install: (app) => {
    register(app)
  }
}
